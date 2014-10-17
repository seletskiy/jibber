{{if .comment}}
    {{template "issue-updated-comment.tpl" .}}
{{end}}

{{if .changelog}}
    {{template "issue-updated-changelog.tpl" .}}
{{end}}
