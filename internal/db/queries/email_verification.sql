-- name: InsertEmailVerificationToken :one
INSERT INTO email_verification_tokens (
    user_id, 
    token, 
    expires_at
) VALUES (
    @user_id, 
    @token,
    @expires_at
) RETURNING *;


-- name: GetEmailVerificationToken :one
SELECT 
    evt.*,
    u.email AS email,
    u.name AS name
FROM email_verification_tokens evt 
JOIN users u ON evt.user_id = u.id
WHERE evt.token = @token;


-- name: DeleteEmailVerificationToken :exec
DELETE FROM email_verification_tokens WHERE id = @token_id;