SELECT
  ed.id
, ed.url
, ed.people
, ed.companies
, ed.created_at
FROM extracted_data ed
WHERE TRUE
{{if .URL -}}
 AND ed.url = @url
{{end -}}
{{if not .CreatedAtFrom.IsZero -}}
 AND ed.created_at >= @created_at_from
{{end -}}
{{if not .CreatedAtTo.IsZero -}}
 AND ed.created_at <= @created_at_to
{{end -}}
ORDER BY ed.created_at DESC
{{if .Limit -}}
 LIMIT @limit
{{end -}}
{{if .Offset -}}
 OFFSET @offset
{{end -}}
