{{if eq .webhookEvent "jira:issue_updated"}}
    {{template "issue-updated-common.tpl" .}}
{{end}}
