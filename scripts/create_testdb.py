#!/usr/bin/env python

import rethinkdb as r


class DatabaseError(Exception):
    pass


class DataFile(object):
    def __init__(self, name, owner, checksum, mediatype, size):
        self.name = name
        self.owner = owner
        self._type = "datafile"
        self.atime = r.now()
        self.birthtime = self.atime
        self.mtime = self.atime
        self.checksum = checksum
        self.current = True
        self.description = self.name
        self.mediatype = mediatype
        self.parent = ""
        self.size = size
        self.uploaded = self.size
        self.usesid = ""


class Project(object):
    def __init__(self, name, owner):
        self._type = "project"
        self.name = name
        self.owner = owner
        self.description = name
        self.birthtime = r.now()
        self.mtime = self.birthtime
        self.size = 0
        self.mediatypes = {}


class DataDir(object):
    def __init__(self, name, owner, parent):
        self._type = "datadir"
        self.name = name
        self.owner = owner
        self.birthtime = r.now()
        self.mtime = self.birthtime
        self.parent = parent


class User(object):
    def __init__(self, name, apikey):
        self._type = "user"
        self.id = name
        self.fullname = name
        self.apikey = apikey
        self.last_login = r.now()
        self.birthtime = self.last_login
        self.mtime = self.birthtime
        self.email = self.id
        self.password = ""
        self.preferences = {
            "tags": [],
            "templates": []
        }


def create_table(table, conn, *args):
    run(r.table_create(table), conn)
    for index_name in args:
        create_index(table, index_name, conn)


def create_index(table, name, conn):
    run(r.table(table).index_create(name), conn)


def run(rql, conn):
    try:
        rql.run(conn)
    except r.RqlRuntimeError:
        pass


def insert(item, table, conn):
    rv = r.table(table).insert(item, return_changes=True).run(conn)
    if rv['inserted'] == 1 or rv['replaced'] == 1:
        if 'changes' in rv:
            return rv['changes'][0]['new_val']
        else:
            return rv['new_val']
    raise DatabaseError()


def make_tables(conn):
    print "Creating tables..."
    create_table("projects", conn, "name", "owner")
    create_table("project2datadir", conn, "datadir_id", "project_id")
    create_table("datadirs", conn, "name", "project")
    create_table("datafiles", conn, "name", "owner", "checksum",
                 "usesid", "mediatype")
    create_table("project2datafile", conn, "project_id", "datafile_id")
    create_table("datadir2datafile", conn, "datadir_id", "datafile_id")
    create_table("users", conn, "apikey")
    print "Done..."


def load_tables(conn):
    print "Loading tables..."
    user = User("test@mc.org", "test")
    insert(user.__dict__, "users", conn)

    project = Project("test", "test@mc.org")
    project.id = "test"
    created_project = insert(project.__dict__, "projects", conn)
    project_id = created_project['id']

    ddir = DataDir("test", "test@mc.org", "")
    ddir.id = "test"
    created_ddir = insert(ddir.__dict__, "datadirs", conn)
    ddir_id = created_ddir['id']

    project2datadir = {
        "project_id": project_id,
        "datadir_id": ddir_id
    }
    insert(project2datadir, "project2datadir", conn)

    # Create a subdirectory
    ddir = DataDir("test/test2", "test@mc.org", "test")
    ddir.id = "test/test2"
    created_ddir = insert(ddir.__dict__, "datadirs", conn)
    ddir_id = created_ddir['id']
    project2datadir = {
        "project_id": project_id,
        "datadir_id": ddir_id
    }
    insert(project2datadir, "project2datadir", conn)

    dfile = DataFile("testfile.txt", "test@mc.org", "abc123", {
        "description": "Text",
        "mime": "text/plain",
        "mime_description": "ASCII Text"
    }, 10)
    dfile.id = "testfile.txt"
    created_dfile = insert(dfile.__dict__, "datafiles", conn)
    dfile_id = created_dfile['id']

    datadir2datafile = {
        "datadir_id": ddir_id,
        "datafile_id": dfile_id
    }
    insert(datadir2datafile, "datadir2datafile", conn)

    project2datafile = {
        "project_id": project_id,
        "datafile_id": dfile_id
    }
    insert(project2datafile, "project2datafile", conn)
    print "Done..."


def create_db():
    print "Creating mctestdb"
    conn = r.connect("localhost", 30815)
    run(r.db_drop("mctestdb"), conn)
    run(r.db_create("mctestdb"), conn)
    conn.close()
    print "Done..."


def main():
    create_db()
    conn = r.connect("localhost", 30815, db="mctestdb")
    make_tables(conn)
    load_tables(conn)


if __name__ == "__main__":
    main()
