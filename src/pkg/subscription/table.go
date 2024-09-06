package subscription

const createTable = `
CREATE TABLE IF NOT EXISTS users (
		user_id INTEGER NOT NULL PRIMARY KEY,
		user TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS subscriptions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		plan_code TEXT NOT NULL,
		datacenters TEXT,
		region TEXT NOT NULL,
		last_check TEXT NOT NULL,
		notifications INTEGER DEFAULT 0,
		UNIQUE (plan_code, datacenters, region)
);

CREATE TABLE IF NOT EXISTS user_subscriptions (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		subscription_id INTEGER NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(user_id),
		FOREIGN KEY(subscription_id) REFERENCES subscriptions(id),
		UNIQUE (user_id, subscription_id)
);
`
