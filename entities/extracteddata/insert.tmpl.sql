INSERT INTO extracted_data (
  url
, people
, companies
, created_at
)
VALUES (
  @url
, @people
, @companies
, @created_at
)
returning id
