-- name: GetChats :many
WITH chat_messages AS (
    SELECT 
        COALESCE(m.group_id, CASE WHEN m.sender_id = @user_id THEN m.recipient_id ELSE m.sender_id END) AS chat_id, 
        (m.group_id IS NOT NULL)::BOOLEAN AS is_group_chat, 
        MAX(m.sent_at) AS last_message_time
    FROM messages m
    WHERE (m.sender_id = @user_id OR m.recipient_id = @user_id OR m.group_id IN (
        SELECT gm.group_id FROM group_members gm WHERE gm.user_id = @user_id
    ))
    GROUP BY chat_id, is_group_chat
),
last_message AS (
    SELECT 
        COALESCE(m.group_id, CASE WHEN m.sender_id = @user_id THEN m.recipient_id ELSE m.sender_id END) AS chat_id, 
        m.id AS message_id, 
        m.sender_id,
        m.recipient_id,
        m.group_id,
        m.text_content,
        m.media,
        m.sent_at
    FROM messages m
    WHERE (m.sender_id = @user_id OR m.recipient_id = @user_id OR m.group_id IN (
        SELECT gm.group_id FROM group_members gm WHERE gm.user_id = @user_id
    ))
),
unread_count AS (
    SELECT 
        COALESCE(m.group_id, CASE WHEN m.sender_id = @user_id THEN m.recipient_id ELSE m.sender_id END) AS chat_id, 
        COUNT(m.id) AS unread_messages_count
    FROM messages m
    LEFT JOIN message_read_receipts mrr ON m.id = mrr.message_id AND mrr.user_id = @user_id
    WHERE m.sender_id <> @user_id 
      AND (m.recipient_id = @user_id OR m.group_id IN (
        SELECT gm.group_id FROM group_members gm WHERE gm.user_id = @user_id
      )) 
      AND mrr.message_id IS NULL
    GROUP BY chat_id
)
SELECT 
    cm.chat_id,
    cm.is_group_chat,
    COALESCE(g.name, u.name) AS chat_name,
    lm.message_id AS last_message_id,
    COALESCE(uc.unread_messages_count, 0) AS unread_message_count
FROM chat_messages cm
JOIN last_message lm ON cm.chat_id = lm.chat_id AND cm.last_message_time = lm.sent_at
LEFT JOIN groups g ON cm.is_group_chat AND cm.chat_id = g.id
LEFT JOIN users u ON NOT cm.is_group_chat AND cm.chat_id = u.id
LEFT JOIN unread_count uc ON cm.chat_id = uc.chat_id
ORDER BY lm.sent_at DESC;



-- name: GetMessages :many
SELECT 
    m.* 
FROM messages m 
WHERE
    CASE 
        WHEN sqlc.narg('target_user_id')::BIGINT IS NOT NULL THEN
            (m.sender_id = @current_user_id::BIGINT OR m.recipient_id = @current_user_id::BIGINT)
            AND
            (m.sender_id = sqlc.narg('target_user_id')::BIGINT OR m.recipient_id = sqlc.narg('target_user_id')::BIGINT)
        WHEN sqlc.narg('target_group_id')::BIGINT IS NOT NULL THEN
            m.group_id = sqlc.narg('target_group_id')::BIGINT
    END
    AND
    (sqlc.narg('cursor_id')::BIGINT IS NULL OR m.id < sqlc.narg('cursor_id')::BIGINT)
ORDER BY m.id DESC
LIMIT @result_limit;


-- name: CheckMessagesHasNextPage :one
SELECT EXISTS(
    SELECT 1 FROM messages m 
    WHERE
        CASE 
            WHEN sqlc.narg('target_user_id')::BIGINT IS NOT NULL THEN
                (m.sender_id = @current_user_id::BIGINT OR m.recipient_id = @current_user_id::BIGINT)
                AND
                (m.sender_id = sqlc.narg('target_user_id')::BIGINT OR m.recipient_id = sqlc.narg('target_user_id')::BIGINT)
            WHEN sqlc.narg('target_group_id')::BIGINT IS NOT NULL THEN
                m.group_id = sqlc.narg('target_group_id')::BIGINT
        END
        AND
        (m.id < @cursor_id::BIGINT)
);


-- name: CheckMessagesHasPreviousPage :one
SELECT EXISTS(
    SELECT 1 FROM messages m 
    WHERE
        CASE 
            WHEN sqlc.narg('target_user_id')::BIGINT IS NOT NULL THEN
                (m.sender_id = @current_user_id::BIGINT OR m.recipient_id = @current_user_id::BIGINT)
                AND
                (m.sender_id = sqlc.narg('target_user_id')::BIGINT OR m.recipient_id = sqlc.narg('target_user_id')::BIGINT)
            WHEN sqlc.narg('target_group_id')::BIGINT IS NOT NULL THEN
                m.group_id = sqlc.narg('target_group_id')::BIGINT
        END
        AND
        (m.id > @cursor_id::BIGINT)
);


-- name: GetBatchedMessages :many
SELECT * FROM messages WHERE id = ANY(@message_ids::BIGINT[]);


-- name: InsertMessage :one
INSERT INTO messages (
    sender_id,
    recipient_id,
    group_id,
    message_type,
    text_content,
    media,
    location,
    reply_for_message_id
) VALUES (
    @sender_id,
    @recipient_id,
    @group_id,
    @message_type,
    @text_content,
    @media,
    @location,
    @reply_for_message_id
) RETURNING *;