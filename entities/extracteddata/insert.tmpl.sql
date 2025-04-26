INSERT INTO extracted_data (
  url
, people
, companies
, created_at
, updated_at
)
VALUES (
  @url
, @people
, @companies
, @created_at
, @updated_at
)
returning id
