-- name: GetGroupByID :one
SELECT * FROM groups WHERE id = @group_id;


-- name: GetBatchedGroups :many
SELECT * FROM groups WHERE id = ANY(@group_ids::BIGINT[]);


-- name: GetGroupMembers :many
SELECT * FROM group_members WHERE group_id = @group_id;


-- name: GetBatchedGroupMembers :many
SELECT 
    gm.*,
    (g.created_by = gm.user_id) AS is_owner
FROM group_members gm
JOIN groups g ON g.id = gm.group_id
WHERE gm.group_id = ANY(@group_ids::BIGINT[])
ORDER BY
    is_owner DESC,
    gm.joined_at ASC;