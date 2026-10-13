package commands

import (
	"bufio"
	"fmt"
	"github.com/google/uuid"
	"github.com/hugojanuario/devdocs/cli/client"
	"github.com/hugojanuario/devdocs/cli/config"
	"github.com/hugojanuario/devdocs/cli/output"
	"github.com/hugojanuario/devdocs/internal/bookmarks"
	"github.com/hugojanuario/devdocs/internal/model"
	"github.com/hugojanuario/devdocs/internal/search"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"io"
	"net/url"
	"os"
	"strings"
	"text/tabwriter"
)

func New() *cobra.Command {
	var format, color string
	var noColor bool
	root := &cobra.Command{Use: "devdocs", Short: "Pesquisa técnica baseada em documentação", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().StringVar(&format, "format", "text", "text, markdown ou json")
	root.PersistentFlags().StringVar(&color, "color", "auto", "auto, always ou never")
	root.PersistentFlags().BoolVar(&noColor, "no-color", false, "Desativar cores")
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if format != "text" && format != "markdown" && format != "json" {
			return fmt.Errorf("formato inválido")
		}
		if color != "auto" && color != "always" && color != "never" {
			return fmt.Errorf("cor inválida")
		}
		return nil
	}
	apiClient := func() (*client.Client, error) {
		c, e := config.Load()
		if e != nil {
			return nil, e
		}
		return client.New(c)
	}
	colored := func(cmd *cobra.Command) bool {
		if noColor || color == "never" || os.Getenv("NO_COLOR") != "" {
			return false
		}
		f, ok := cmd.OutOrStdout().(*os.File)
		return ok && term.IsTerminal(int(f.Fd())) && (color == "always" || color == "auto")
	}
	conf := &cobra.Command{Use: "config"}
	var trusted bool
	setServer := &cobra.Command{Use: "set-server <url>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if e := config.ValidateServer(args[0], trusted); e != nil {
			return e
		}
		c, e := config.Load()
		if e != nil {
			return e
		}
		c.Server = args[0]
		c.TrustedPrivateNetwork = trusted
		return config.Save(c)
	}}
	setServer.Flags().BoolVar(&trusted, "trusted-private-network", false, "Autorizar HTTP apenas numa rede privada confiável")
	setToken := &cobra.Command{Use: "set-token", Short: "Ler token sem eco; aceita stdin redirecionado", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		var b []byte
		var e error
		f, ok := cmd.InOrStdin().(*os.File)
		if ok && term.IsTerminal(int(f.Fd())) {
			fmt.Fprint(cmd.ErrOrStderr(), "Token: ")
			b, e = term.ReadPassword(int(f.Fd()))
			fmt.Fprintln(cmd.ErrOrStderr())
		} else {
			reader := bufio.NewReader(io.LimitReader(cmd.InOrStdin(), 4097))
			b, e = io.ReadAll(reader)
		}
		if e != nil {
			return e
		}
		token := strings.TrimSpace(string(b))
		if len(token) < 32 || len(token) > 4096 || strings.ContainsAny(token, "\r\n") {
			return fmt.Errorf("token deve ter entre 32 e 4096 caracteres")
		}
		c, e := config.Load()
		if e != nil {
			return e
		}
		c.Token = token
		return config.Save(c)
	}}
	conf.AddCommand(setServer, setToken)
	root.AddCommand(conf)
	var save, ai bool
	var language string
	var sources []string
	query := &cobra.Command{Use: "search <pergunta>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		c, e := apiClient()
		if e != nil {
			return e
		}
		var a model.Answer
		if e = c.Do(cmd.Context(), "POST", "/search", model.SearchRequest{Query: args[0], Language: language, PreferredSources: sources, AIEnabled: ai}, &a); e != nil {
			return e
		}
		if save {
			var v bookmarks.Saved
			if e = c.Do(cmd.Context(), "POST", "/saved", map[string]any{"response": a}, &v); e != nil {
				return e
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "Salvo:", v.ID)
		}
		return output.Render(cmd.OutOrStdout(), a, format, colored(cmd))
	}}
	query.Flags().BoolVar(&save, "save", false, "Persistir explicitamente a resposta")
	query.Flags().BoolVar(&ai, "ai", false, "Explicar trechos com IA opcional configurada no servidor")
	query.Flags().StringVar(&language, "language", "pt-BR", "pt-BR ou en")
	query.Flags().StringSliceVar(&sources, "sources", nil, "official, complementary")
	root.AddCommand(query)
	saved := &cobra.Command{Use: "saved"}
	var page, size int
	var filter, technology string
	list := &cobra.Command{Use: "list", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		c, e := apiClient()
		if e != nil {
			return e
		}
		q := url.Values{"page": {fmt.Sprint(page)}, "page_size": {fmt.Sprint(size)}, "q": {filter}, "technology": {technology}}
		var p bookmarks.Page
		if e = c.Do(cmd.Context(), "GET", "/saved?"+q.Encode(), nil, &p); e != nil {
			return e
		}
		if format == "json" {
			return output.JSON(cmd.OutOrStdout(), p)
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTÍTULO\tTECNOLOGIA")
		for _, v := range p.Items {
			fmt.Fprintf(w, "%s\t%s\t%s\n", v.ID, oneLine(v.Title), oneLine(v.Technology))
		}
		_ = w.Flush()
		fmt.Fprintf(cmd.OutOrStdout(), "Página %d | total %d\n", p.Page, p.Total)
		return nil
	}}
	list.Flags().IntVar(&page, "page", 1, "Página")
	list.Flags().IntVar(&size, "page-size", 20, "Itens por página")
	list.Flags().StringVar(&filter, "query", "", "Filtrar título/pergunta")
	list.Flags().StringVar(&technology, "technology", "", "Filtrar tecnologia")
	saved.AddCommand(list)
	validID := func(args []string) error {
		if _, e := uuid.Parse(args[0]); e != nil {
			return fmt.Errorf("UUID inválido")
		}
		return nil
	}
	saved.AddCommand(&cobra.Command{Use: "show <uuid>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if e := validID(args); e != nil {
			return e
		}
		c, e := apiClient()
		if e != nil {
			return e
		}
		var v bookmarks.Saved
		if e = c.Do(cmd.Context(), "GET", "/saved/"+args[0], nil, &v); e != nil {
			return e
		}
		if format == "json" {
			return output.JSON(cmd.OutOrStdout(), v)
		}
		answer := v.Response
		answer.Title = v.Title
		return output.Render(cmd.OutOrStdout(), answer, format, colored(cmd))
	}})
	saved.AddCommand(&cobra.Command{Use: "rename <uuid> <title>", Args: cobra.ExactArgs(2), RunE: func(cmd *cobra.Command, args []string) error {
		if e := validID(args); e != nil {
			return e
		}
		c, e := apiClient()
		if e != nil {
			return e
		}
		var v bookmarks.Saved
		if e = c.Do(cmd.Context(), "PATCH", "/saved/"+args[0], map[string]string{"title": args[1]}, &v); e != nil {
			return e
		}
		if format == "json" {
			return output.JSON(cmd.OutOrStdout(), v)
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Renomeado:", oneLine(v.Title))
		return nil
	}})
	saved.AddCommand(&cobra.Command{Use: "delete <uuid>", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		if e := validID(args); e != nil {
			return e
		}
		c, e := apiClient()
		if e != nil {
			return e
		}
		if e = c.Do(cmd.Context(), "DELETE", "/saved/"+args[0], nil, nil); e != nil {
			return e
		}
		if format == "json" {
			return output.JSON(cmd.OutOrStdout(), map[string]bool{"deleted": true})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "Excluído:", args[0])
		return nil
	}})
	root.AddCommand(saved)
	root.AddCommand(&cobra.Command{Use: "technologies", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		c, e := apiClient()
		if e != nil {
			return e
		}
		var v []search.Technology
		if e = c.Do(cmd.Context(), "GET", "/technologies", nil, &v); e != nil {
			return e
		}
		if format == "json" {
			return output.JSON(cmd.OutOrStdout(), v)
		}
		for _, t := range v {
			fmt.Fprintf(cmd.OutOrStdout(), "%s: %s\n", t.ID, t.Name)
		}
		return nil
	}})
	root.AddCommand(&cobra.Command{Use: "ping", Args: cobra.NoArgs, RunE: func(cmd *cobra.Command, args []string) error {
		c, e := apiClient()
		if e != nil {
			return e
		}
		var v map[string]string
		if e = c.Do(cmd.Context(), "GET", "/ready", nil, &v); e != nil {
			return e
		}
		if format == "json" {
			return output.JSON(cmd.OutOrStdout(), v)
		}
		fmt.Fprintln(cmd.OutOrStdout(), v["status"])
		return nil
	}})
	return root
}
func oneLine(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 32 || r == 127 || (r >= 128 && r <= 159) {
			return ' '
		}
		return r
	}, s)
}
