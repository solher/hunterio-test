SELECT
  ed.id
, ed.url
, ed.people
, ed.companies
, ed.created_at
, ed.updated_at
FROM extracted_data ed
WHERE TRUE
{{if .URL -}}
 AND ed.url = @url
{{end -}}
