CREATE TABLE organizations (
	id INTEGER NOT NULL,
	name VARCHAR(40),
	contact VARCHAR,
	comments VARCHAR,
	PRIMARY KEY (id)
);
CREATE TABLE features (
	feat VARCHAR(16) NOT NULL,
	ispackage BOOLEAN,
	description VARCHAR,
	PRIMARY KEY (feat),
	CHECK (ispackage IN (0, 1))
);
CREATE TABLE keys (
	id VARCHAR(10) NOT NULL,
	assigned_org INTEGER,
	comments VARCHAR,
	PRIMARY KEY (id),
	FOREIGN KEY(assigned_org) REFERENCES organizations (id)
);
CREATE TABLE pkgcontent (
	pkg VARCHAR(16) NOT NULL,
	feat VARCHAR(16) NOT NULL,
	PRIMARY KEY (pkg, feat),
	FOREIGN KEY(pkg) REFERENCES features (feat),
	FOREIGN KEY(feat) REFERENCES features (feat)
);
CREATE TABLE templates (
	name VARCHAR(10) NOT NULL,
	feat VARCHAR(16) NOT NULL,
	ver FLOAT,
	dup VARCHAR(4),
	PRIMARY KEY (name, feat),
	FOREIGN KEY(feat) REFERENCES features (feat)
);
CREATE TABLE licensesets (
	keyid VARCHAR(10) NOT NULL,
	feat VARCHAR(16) NOT NULL,
	ver FLOAT,
	dup VARCHAR(4),
	count INTEGER,
	start DATE,
	"end" DATE,
	PRIMARY KEY (keyid, feat),
	FOREIGN KEY(keyid) REFERENCES keys (id),
	FOREIGN KEY(feat) REFERENCES features (feat)
);
CREATE TABLE history (
	id INTEGER NOT NULL,
	orgname VARCHAR(40),
	whenissued TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	xml VARCHAR,
	PRIMARY KEY (id)
);