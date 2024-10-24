CREATE EXTENSION postgis;


CREATE TABLE roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    description TEXT,
    sort_index BIGINT NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    CONSTRAINT roles_unique_name UNIQUE (name)
);


CREATE TABLE permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    description TEXT,
    sort_index BIGINT NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    CONSTRAINT permissions_unique_name UNIQUE (name)
);



CREATE TABLE role_permissions (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    role_id BIGINT NOT NULL,
    permission_id BIGINT,

    PRIMARY KEY (id),
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions (id) ON DELETE CASCADE,
    CONSTRAINT role_permissions_unique_role_permission UNIQUE (role_id, permission_id)
);




CREATE TABLE users (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    username TEXT NOT NULL,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    password_hash BYTEA NOT NULL,
    bio TEXT,
    image TEXT,
    online BOOLEAN NOT NULL DEFAULT FALSE,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    PRIMARY KEY (id),
    CONSTRAINT users_unique_username UNIQUE (username),
    CONSTRAINT users_unique_email UNIQUE (email)
);



CREATE TABLE user_roles (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles (id) ON DELETE CASCADE,
    CONSTRAINT user_roles_unique_user_role UNIQUE (user_id, role_id)
);




CREATE TABLE refresh_tokens (
    id UUID,
    user_id BIGINT NOT NULL,
    token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    issued_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);




CREATE TABLE user_hierarchy (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    parent_id BIGINT NOT NULL,
    child_id BIGINT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (parent_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (child_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT user_hierarchy_unique_parent_and_child UNIQUE (parent_id, child_id),
    CONSTRAINT check_user_hierarchy_inequal_parent_and_child CHECK (parent_id <> child_id),
    CONSTRAINT check_user_hierarchy_no_circular_relationship CHECK (parent_id < child_id)
);



CREATE TABLE email_verification_tokens (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    token TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,

    
    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT email_verification_tokens_unique_token UNIQUE (token),
    CONSTRAINT email_verification_tokens_unique_user_token UNIQUE (user_id, token)
);



CREATE TABLE friendships (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    user_id BIGINT NOT NULL,
    friend_id BIGINT NOT NULL,
    status TEXT NOT NULL, -- e.g., 'pending', 'accepted', 'blocked'
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    FOREIGN KEY (friend_id) REFERENCES users (id) ON DELETE CASCADE
);



CREATE TABLE groups (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    image TEXT,
    description TEXT,
    created_by BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (created_by) REFERENCES users (id) ON DELETE SET NULL
);



CREATE TABLE group_members (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    group_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    joined_at TIMESTAMPTZ DEFAULT NOW(),
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,

    PRIMARY KEY (id),
    FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT group_members_unique_group_user UNIQUE (group_id, user_id)
);




CREATE TABLE messages (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    sender_id BIGINT NOT NULL,   
    recipient_id BIGINT, -- recipient id will be null if the message is sent to a group              
    group_id BIGINT, -- group id will be null if the message is a direct message           
    message_type TEXT NOT NULL, -- Type of message: 'system', 'text', 'audio', 'video', 'document', 'location',         
    text_content TEXT,                  
    media TEXT,
    location GEOGRAPHY(POINT), -- Location for location and live location messages.
    reply_for_message_id BIGINT, -- Will be not null if a message is a reply for another message
    sent_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    deleted_by BIGINT,

    PRIMARY KEY (id),
    FOREIGN KEY (sender_id) REFERENCES users (id),
    FOREIGN KEY (recipient_id) REFERENCES users (id),
    FOREIGN KEY (group_id) REFERENCES groups (id) ON DELETE CASCADE,
    FOREIGN KEY (reply_for_message_id) REFERENCES messages (id) ON DELETE CASCADE,
    FOREIGN KEY (deleted_by) REFERENCES users (id)
);




CREATE TABLE message_reactions (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    message_id INT NOT NULL,                    
    user_id INT NOT NULL,                
    emoji TEXT NOT NULL,                 
    reacted_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT message_reactions_unique_message_user UNIQUE (message_id, user_id)
);




CREATE TABLE message_read_receipts (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    message_id BIGINT NOT NULL,                     
    user_id BIGINT NOT NULL,                        
    read_at TIMESTAMPTZ NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (message_id) REFERENCES messages (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT message_read_receipts_unique_message_user UNIQUE (message_id, user_id)
);



CREATE TABLE posts (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    caption TEXT,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    sort_index INT NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);



CREATE TABLE post_media (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    post_id BIGINT NOT NULL,
    media TEXT NOT NULL,
    sort_index INT NOT NULL DEFAULT 0,

    PRIMARY KEY (id),
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE
);



CREATE TABLE post_likes (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    post_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    liked_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);



CREATE TABLE post_comments (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    post_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    parent_id BIGINT,

    PRIMARY KEY (id),
    FOREIGN KEY (post_id) REFERENCES posts (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL,
    FOREIGN KEY (parent_id) REFERENCES post_comments (id) ON DELETE CASCADE
);



CREATE TABLE post_comment_likes (
    id BIGINT GENERATED ALWAYS AS IDENTITY,
    post_comment_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    liked_at TIMESTAMPTZ DEFAULT NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (post_comment_id) REFERENCES post_comments (id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);





