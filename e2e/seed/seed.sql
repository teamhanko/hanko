INSERT INTO users
(id, created_at, updated_at)
VALUES
('357461f1-458a-42c8-abf3-05eabfc36ffd', current_timestamp, current_timestamp);

INSERT INTO emails
(id, user_id, address, verified, created_at, updated_at)
VALUES
('47c082da-b70a-4ccc-bc5f-1481b3499273', '357461f1-458a-42c8-abf3-05eabfc36ffd', 'test.verified@example.com', true, current_timestamp, current_timestamp);

INSERT INTO primary_emails
(id, email_id, user_id, created_at, updated_at)
VALUES
('8de035cd-3d21-415c-8844-644fe40d7d74', '47c082da-b70a-4ccc-bc5f-1481b3499273', '357461f1-458a-42c8-abf3-05eabfc36ffd', current_timestamp, current_timestamp);


INSERT INTO users
(id, created_at, updated_at)
VALUES
('92789b5f-d3ad-46bd-93e8-42afaa4e15ff', current_timestamp, current_timestamp);

INSERT INTO emails
(id, user_id, address, verified, created_at, updated_at)
VALUES
('772663eb-e598-4a63-88df-ceef1a329833', '92789b5f-d3ad-46bd-93e8-42afaa4e15ff', 'test.unverified@example.com', false, current_timestamp, current_timestamp);

INSERT INTO primary_emails
(id, email_id, user_id, created_at, updated_at)
VALUES
('36b9ad84-ff0b-4f9d-af5e-465a5766d071', '772663eb-e598-4a63-88df-ceef1a329833', '92789b5f-d3ad-46bd-93e8-42afaa4e15ff', current_timestamp, current_timestamp);
