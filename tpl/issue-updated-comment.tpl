{{template "head.tpl" .}}
{{"Commented:" | indent 2}}{{"\n"}}
{{.comment.body | indent 3}}{{"\n"}}
{{template "footer.tpl" .}}
