package search

import (
	"net/url"
	"regexp"
	"strings"
)

type Technology struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Aliases []string `json:"aliases"`
}
type Source struct {
	Domain     string `json:"domain"`
	PathPrefix string `json:"path_prefix"`
	Technology string `json:"technology"`
	Official   bool   `json:"official"`
	IndexURL   string `json:"index_url"`
}

var Technologies = []Technology{
	{"python", "Python", []string{"python", "dotenv", ".env"}}, {"pandas", "Pandas", []string{"pandas", "dataframe", "read_csv"}}, {"postgresql", "PostgreSQL", []string{"postgresql", "postgres", "cte", "sql", "join"}}, {"go", "Go", []string{"golang", "go"}}, {"docker", "Docker", []string{"docker", "container", "volumes"}}, {"airflow", "Airflow", []string{"airflow", "dag"}}, {"spark", "Spark", []string{"spark", "pyspark"}}, {"dbt", "dbt", []string{"dbt"}}, {"fastapi", "FastAPI", []string{"fastapi"}}, {"kubernetes", "Kubernetes", []string{"kubernetes", "k8s", "kubectl"}}, {"sqlalchemy", "SQLAlchemy", []string{"sqlalchemy"}},
}
var Sources = []Source{
	{"bbc2.github.io", "/python-dotenv/", "python", true, ""},
	{"docs.python.org", "/", "python", true, "https://docs.python.org/3/searchindex.js"},
	{"pandas.pydata.org", "/", "pandas", true, "https://pandas.pydata.org/docs/searchindex.js"},
	{"www.postgresql.org", "/docs/", "postgresql", true, ""}, {"postgresql.org", "/docs/", "postgresql", true, ""},
	{"go.dev", "/doc/", "go", true, ""}, {"pkg.go.dev", "/", "go", true, ""}, {"docs.docker.com", "/", "docker", true, ""},
	{"airflow.apache.org", "/docs/", "airflow", true, "https://airflow.apache.org/docs/apache-airflow/stable/searchindex.js"},
	{"spark.apache.org", "/docs/", "spark", true, ""}, {"docs.getdbt.com", "/", "dbt", true, ""}, {"fastapi.tiangolo.com", "/", "fastapi", true, ""}, {"kubernetes.io", "/docs/", "kubernetes", true, ""}, {"docs.sqlalchemy.org", "/", "sqlalchemy", true, "https://docs.sqlalchemy.org/en/20/searchindex.js"},
}

func Classify(raw string) (Source, bool) {
	u, e := url.Parse(raw)
	if e != nil {
		return Source{}, false
	}
	for _, s := range Sources {
		if strings.EqualFold(u.Hostname(), s.Domain) && strings.HasPrefix(u.Path, s.PathPrefix) {
			return s, true
		}
	}
	return Source{Domain: u.Hostname()}, false
}

var tokenRE = regexp.MustCompile(`[\pL\pN_.]+`)

func Detect(q string) []string {
	words := tokenRE.FindAllString(strings.ToLower(q), -1)
	out := []string{}
	for _, t := range Technologies {
		found := false
		for _, a := range t.Aliases {
			for _, w := range words {
				if w == a {
					found = true
				}
			}
		}
		if found {
			out = append(out, t.ID)
		}
	}
	return out
}
func Terms(q string) []string {
	q = strings.ToLower(q)
	replacements := map[string]string{"carregar": "load", "escrever": "write", "subir": "up", "serviços": "services", "servicos": "services", ".env": "dotenv", "listar": "list", "salvar": "save", "ler": "read", "leitura": "read", "csv": "csv", "arquivos": "file", "arquivo": "file", "volumes": "volume", "criar": "create", "utilizar": "use", "usar": "use", "três": "three", "juncao": "join", "junção": "join", "ambiente": "environment"}
	stop := map[string]bool{"e": true, "para": true, "por": true, "que": true, "qual": true, "quero": true, "preciso": true, "fazer": true, "of": true, "with": true, "an": true, "please": true, "como": true, "com": true, "no": true, "na": true, "em": true, "de": true, "do": true, "da": true, "o": true, "a": true, "um": true, "uma": true, "usando": true, "how": true, "to": true, "in": true, "the": true, "use": true, "using": true}
	seen := map[string]bool{}
	out := []string{}
	for _, w := range tokenRE.FindAllString(q, -1) {
		if v, ok := replacements[w]; ok {
			w = v
		}
		if !stop[w] && !seen[w] {
			out = append(out, w)
			seen[w] = true
			if w == "cte" && !seen["with"] {
				out = append(out, "with")
				seen["with"] = true
			}
		}
	}
	return out
}

var versionRE = regexp.MustCompile(`(?:\b|/)(\d+\.\d+(?:\.\d+)?)(?:\b|/)`)

func Version(s string) string { return versionRE.FindString(s) }

// Technology names identify the collection, not the subject of the question.
// Counting "Python" as evidence makes unrelated Python pages look relevant.
func TopicTerms(q string, technologies []string) []string {
	terms := []string{}
	for _, term := range Terms(q) {
		if !contains(technologies, term) && term != "fazer" {
			terms = append(terms, term)
		}
	}
	return terms
}
