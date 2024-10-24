-- name: GetUser :one
SELECT 
    u.*, 
    COUNT(f.id) AS friend_count
FROM 
    users u
LEFT JOIN 
    friendships f
    ON (u.id = f.user_id OR u.id = f.friend_id) AND f.status = 'accepted'
WHERE 
    u.id = @user_id
GROUP BY 
    u.id;


-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = @email;


-- name: GetBatchedUsers :many
SELECT 
    u.*, 
    COUNT(f.id) AS friend_count
FROM 
    users u
LEFT JOIN 
    friendships f
    ON (u.id = f.user_id OR u.id = f.friend_id) AND f.status = 'accepted'
WHERE 
    u.id = ANY(@user_ids::BIGINT[])
GROUP BY 
    u.id;


-- name: CheckUsernameExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE username = @username);


-- name: CheckEmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = @email);



-- name: InsertUser :one
INSERT INTO users (
    username,
    email,
    name,
    password_hash,
    bio,
    image
) VALUES (
    @username,
    @email,
    @name,
    @password_hash,
    @bio,
    @image
) RETURNING *;


-- name: InsertUserRole :exec
INSERT INTO user_roles (
    user_id,
    role_id
) VALUES (
    @user_id,
    (SELECT r.id FROM roles r WHERE r.name = @role_name)
) ON CONFLICT(user_id, role_id) DO NOTHING;


-- name: GetUserRoles :one
SELECT ARRAY_AGG(r.name)::TEXT[] AS roles FROM roles r JOIN user_roles ur ON r.id = ur.role_id WHERE ur.user_id = @user_id GROUP BY ur.user_id;


-- name: UpdateUserEmailVerifiedAt :exec
UPDATE users SET email_verified_at = @email_verified_at WHERE id = @user_id;


-- name: UpdateUser :exec
UPDATE users SET
    username = @username,
    name = @name,
    password_hash = COALESCE(sqlc.narg('password_hash'), password_hash),
    bio = @bio,
    image = @image
WHERE id = @user_id;


-- name: UpdateUserOnlineStatus :exec
UPDATE users SET online = @online WHERE id = @user_id;

