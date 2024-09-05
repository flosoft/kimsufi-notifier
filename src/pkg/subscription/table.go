package subscription

const createTable = `
CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		plan_code TEXT NOT NULL,
		datacenters TEXT,
		user_id INTEGER NOT NULL,
		user TEXT NOT NULL,
		last_check TEXT NOT NULL,
		UNIQUE (id, user_id)
);
`
