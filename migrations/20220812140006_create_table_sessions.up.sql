-- Copyright (c) 2022 Isaque Veras
-- Use of this source code is governed by MIT style
-- license that can be found in the LICENSE file.

CREATE TABLE "sessions" (
	id						UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
	token 				VARCHAR NOT NULL,
	user_id				UUID NOT NULL REFERENCES users (id),
	expires_at		TIMESTAMP WITH TIME ZONE NOT NULL,
	created_at		TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at		TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);