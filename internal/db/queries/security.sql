-- name: UpsertRole :exec
INSERT INTO roles 
    (
        name, 
        description,
        sort_index
    )
VALUES (
    @name, 
    @description,
    @sort_index
)
ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    sort_index = EXCLUDED.sort_index;


-- name: GetRoles :many
SELECT * FROM roles;



-- name: DeleteRole :exec
DELETE FROM roles WHERE name = @name;







-- name: UpsertPermission :exec
INSERT INTO permissions 
    (
        name, 
        description,
        sort_index
    )
VALUES (
    @name, 
    @description,
    @sort_index
)
ON CONFLICT (name) DO UPDATE
SET 
    description = EXCLUDED.description,
    sort_index = EXCLUDED.sort_index;


-- name: GetPermissions :many
SELECT * FROM permissions;



-- name: DeletePermission :exec
DELETE FROM permissions WHERE name = @name;




-- name: DeleteRolePermissions :exec
DELETE FROM role_permissions WHERE role_id = (SELECT r.id FROM roles r WHERE r.name = @role_name);


-- name: InsertRolePermission :exec
INSERT INTO role_permissions (
    role_id,
    permission_id
) VALUES (
    (SELECT r.id FROM roles r WHERE r.name = @role_name),
    (SELECT p.id FROM permissions p WHERE p.name = @permission_name)

);







-- name: InsertRefreshToken :exec
INSERT INTO refresh_tokens (
    id,
    user_id,
    token,
    expires_at,
    issued_at
) VALUES (
    @id,
    @user_id,
    @token,
    @expires_at,
    @issued_at
);


-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE id = @token_id;


-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens WHERE id = @token_id;


-- name: DeleteRefreshTokensByUserID :exec
DELETE FROM refresh_tokens WHERE user_id = @user_id;
