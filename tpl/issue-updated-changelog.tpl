{{with $root := .}}
{{with $change := index .changelog.items 0}}
    {{if eq $change.field "status"}}
        {{template "head.tpl" $root}}
        {{"Moved:" | indent 2}}{{"\n"}}
        {{$change.fromString | indent 3}}{{" -> "}}{{$change.toString}}{{"\n"}}
        {{template "footer.tpl" $root}}
    {{end}}

    {{if eq $change.field "labels"}}
        {{if not (and (hasTag "jwh:in-work" $change.toString) (hasTag "jwh:in-work" $change.fromString))}}
            {{if hasTag "jwh:in-work" $change.toString}}
                {{template "head.tpl" $root}}
                {{"Work on the issue was started." | indent 2}}{{"\n"}}
                {{template "footer.tpl" $root}}
            {{end}}
            {{if hasTag "jwh:in-work" $change.fromString}}
                {{template "head.tpl" $root}}
                {{"Work on the issue was stopped." | indent 2}}{{"\n"}}
                {{template "footer.tpl" $root}}
            {{end}}
        {{end}}
    {{end}}
{{end}}
{{end}}
