{{if eq .webhookEvent "jira:issue_created"}}
    {{template "issue-created-common.tpl" .}}
{{end}}

{{if eq .webhookEvent "jira:issue_updated"}}
    {{template "issue-updated-common.tpl" .}}
{{end}}
