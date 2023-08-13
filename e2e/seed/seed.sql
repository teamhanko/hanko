INSERT INTO users
(id, created_at, updated_at)
VALUES
('357461f1-458a-42c8-abf3-05eabfc36ffd', current_timestamp, current_timestamp);

INSERT INTO emails
(id, user_id, address, verified, created_at, updated_at)
VALUES
('47c082da-b70a-4ccc-bc5f-1481b3499273', '357461f1-458a-42c8-abf3-05eabfc36ffd', 'test@example.com', true, current_timestamp, current_timestamp);

INSERT INTO primary_emails
(id, email_id, user_id, created_at, updated_at)
VALUES
('8de035cd-3d21-415c-8844-644fe40d7d74', '47c082da-b70a-4ccc-bc5f-1481b3499273', '357461f1-458a-42c8-abf3-05eabfc36ffd', current_timestamp, current_timestamp);
