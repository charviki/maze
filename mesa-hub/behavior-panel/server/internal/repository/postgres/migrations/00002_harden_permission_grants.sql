-- +goose Up

-- 同一申请单内的资源动作对只允许生成一条 grant，避免并发审批或重复目标写出重复授权。
ALTER TABLE auth.permission_grants
    ADD CONSTRAINT uq_permission_grants_request_resource_action
    UNIQUE (source_request_id, resource, action);

-- +goose Down

ALTER TABLE auth.permission_grants
    DROP CONSTRAINT IF EXISTS uq_permission_grants_request_resource_action;
