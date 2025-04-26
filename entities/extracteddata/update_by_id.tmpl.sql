UPDATE extracted_data
SET
  people = @people
, companies = @companies
WHERE id = @id
