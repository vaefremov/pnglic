! Legacy schema in sqlite3

CREATE TABLE organizations (id integer primary key AUTOINCREMENT, name varchar(40), contact text, comments text);
CREATE TABLE keys (id varchar(10) primary key not null, assigned_org integer, comments text);
CREATE TABLE history (orgname VARCHAR(40), whenissued VARCHAR(16) DEFAULT CURRENT_TIMESTAMP NOT NULL, xml TEXT);

CREATE TABLE features (feat varchar(16) primary key, ispackage integer default 0, description text);
CREATE TABLE pkgcontent (pkg varchar(10), feat varchar(10), primary key (pkg, feat));

CREATE TABLE licensesets (keyid varchar(10), feat varchar(16), ver float, count integer, start DATE, end DATE, dup varchar(4), primary key (keyid, feat));

CREATE TABLE templates (name varchar(10) not null, feat varchar(16), ver float, dup varchar(4), primary key (name, feat));