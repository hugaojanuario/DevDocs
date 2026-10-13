package extraction

import "testing"

func TestTranscriptSeparatesCommandsFromResults(t *testing.T) {
	tests := []struct{ raw, lang, code, output string }{
		{">>> pd.read_csv(\"data.csv\")\n   Name  Value\n0  foo   1", "python", `pd.read_csv("data.csv")`, "   Name  Value\n0  foo   1"},
		{">>> for x in range(2):\n...     print(x)\n...\n0\n1", "python", "for x in range(2):\n    print(x)", "0\n1"},
		{"$ docker volume ls\nDRIVER VOLUME NAME\nlocal data", "shell", "docker volume ls", "DRIVER VOLUME NAME\nlocal data"},
		{"$ docker run \\\n    --name test alpine\ncontainer-id", "shell", "docker run \\\n    --name test alpine", "container-id"},
		{"if True:\n    print('>>> keep literal')", "python", "if True:\n    print('>>> keep literal')", ""},
		{"SELECT * FROM customers;", "sql", "SELECT * FROM customers;", ""},
		{"In [2]: titanic = pd.read_csv(\"data/titanic.csv\")\nIn [3]: titanic\nOut[3]:\n  Name\n0 foo", "python", "titanic = pd.read_csv(\"data/titanic.csv\")\ntitanic", "  Name\n0 foo"},
		{"In [1]: for x in range(2):\n   ...:     print(x)\n   ...:\n0\n1", "python", "for x in range(2):\n    print(x)", "0\n1"},
	}
	for _, tt := range tests {
		code, output := NormalizeTranscript(tt.raw, tt.lang)
		if code != tt.code || output != tt.output {
			t.Fatalf("%q: code=%q output=%q", tt.raw, code, output)
		}
	}
}

func TestExtractReusableContextAndPreferSimpleCode(t *testing.T) {
	raw := []byte(`<main><h1>read_csv</h1><dl class="py function"><dd><p>Read CSV into a DataFrame.</p></dd></dl><pre class="highlight-python">import pandas as pd</pre><pre class="highlight-pycon">&gt;&gt;&gt; pd.read_csv("data.csv")
   Name Value
0  foo  1</pre><pre class="highlight-pycon">&gt;&gt;&gt; for i in range(3):
...     pd.read_csv("data.csv")</pre></main>`)
	d, err := Extract(raw, []string{"read", "csv"}, "source-1", "pandas")
	if err != nil || len(d.Examples) != 2 {
		t.Fatal(d, err)
	}
	ex := d.Examples[0]
	if ex.Language != "python" || ex.Code != `pd.read_csv("data.csv")` || ex.Setup != "import pandas as pd" || ex.Output != "   Name Value\n0  foo  1" || ex.SourceCode == ex.Code {
		t.Fatal("failed to select and structure the simplest relevant example", ex)
	}
}

func TestNoInventedContext(t *testing.T) {
	if got := ImportContext([]string{"import os"}, "load_dotenv() # sets os.environ"); got != "" {
		t.Fatal("comment treated as dependency", got)
	}
	if got := ImportContext([]string{"print('pandas')", "    import pandas as pd"}, "pd.read_csv('file.csv')"); got != "" {
		t.Fatal("guessed imports", got)
	}
	if got := ImportContext([]string{"In [1]: import pandas as pd"}, "pd.read_csv('file.csv')"); got != "import pandas as pd" {
		t.Fatal("missed IPython context", got)
	}
	if got := ImportContext([]string{"import pandas as pd\nimport unrelated as pd"}, "pd.read_csv('file.csv')"); got != "" {
		t.Fatal("ambiguous alias reused", got)
	}
}

func TestGenericPythonQueryPrefersScriptOverNotebookMagic(t *testing.T) {
	raw := []byte(`<main><h1>python-dotenv</h1><p>Read environment variables from dotenv files.</p><pre>from dotenv import load_dotenv

load_dotenv() # sets os.environ</pre><pre>%load_ext dotenv
%dotenv</pre></main>`)
	d, err := Extract(raw, []string{"load", "dotenv"}, "source-1", "python")
	if err != nil || len(d.Examples) != 2 || d.Examples[0].Code != "from dotenv import load_dotenv\n\nload_dotenv() # sets os.environ" {
		t.Fatal(d, err)
	}
}

func TestPythonQuestionDoesNotSelectINIConfiguration(t *testing.T) {
	raw := []byte(`<main><h1>DAGs</h1><p>Create a DAG using the documented Python definitions.</p><pre class="highlight-default">[core]
min_serialized_dag_update_interval = 30
compress_serialized_dags = False</pre><pre class="highlight-python">with DAG("example_dag") as dag:
    task = EmptyOperator(task_id="example")</pre></main>`)
	d, err := Extract(raw, []string{"create", "dag"}, "source-1", "airflow")
	if err != nil || len(d.Examples) != 2 || d.Examples[0].Language != "python" || d.Examples[1].Language != "ini" {
		t.Fatal(d, err)
	}
}

func TestCreatingSubjectPrefersConstructionOverTesting(t *testing.T) {
	raw := []byte(`<main><h1>Dags</h1><p>A Dag is a collection of tasks with dependencies.</p><p>Create defaults for Operators inside a Dag.</p><pre>dag.test(use_executor=True)</pre><pre>with DAG("example") as dag:
    task = EmptyOperator(task_id="task")</pre></main>`)
	d, err := Extract(raw, []string{"create", "dag"}, "source-1", "airflow")
	if err != nil || len(d.Examples) != 2 || d.Examples[0].Code == "dag.test(use_executor=True)" || d.Summary != "A Dag is a collection of tasks with dependencies." {
		t.Fatal(d, err)
	}
}
