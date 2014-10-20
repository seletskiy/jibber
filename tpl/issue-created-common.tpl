{{template "head.tpl" .}}
{{"Created:" | indent 2}}{{"\n"}}
{{if .issue.fields.description}}
	{{.issue.fields.description | indent 3}}
{{else}}
	{{"<no description>" | indent 3}}
{{end}}
{{"\n"}}
{{template "footer.tpl" .}}
