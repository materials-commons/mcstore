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
        self.project = ""


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
        self.admin = False
        self.preferences = {
            "tags": [],
            "templates": []
        }


class Access(object):
    def __init__(self, project_id, project_name, user):
        now = r.now()
        self.project_id = project_id
        self.project_name = project_name
        self.user_id = user
        self.birthtime = now
        self.mtime = now
        self.status = ""
        self.dataset = ""
        self.permissions = ""


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
    create_table("access", conn, "user_id", "project_id")
    create_table("uploads", conn, "owner", "project_id")
    print "Done..."


def load_tables(conn):
    print "Loading tables..."
    user = User("admin@mc.org", "admin")
    user.admin = True
    insert(user.__dict__, "users", conn)

    user = User("test@mc.org", "test")
    insert(user.__dict__, "users", conn)
    user = User("test1@mc.org", "test1")
    insert(user.__dict__, "users", conn)
    user = User("test2@mc.org", "test2")
    insert(user.__dict__, "users", conn)
    user = User("test3@mc.org", "test3")
    insert(user.__dict__, "users", conn)

    project = Project("test", "test@mc.org")
    project.id = "test"
    created_project = insert(project.__dict__, "projects", conn)
    project_id = created_project['id']

    uaccess = Access("test", "test", "test@mc.org")
    insert(uaccess.__dict__, "access", conn)
    uaccess = Access("test", "test", "test1@mc.org")
    insert(uaccess.__dict__, "access", conn)

    ddir = DataDir("test", "test@mc.org", "")
    ddir.id = "test"
    ddir.project = project_id
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
    ddir.project = project_id
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

    project = Project("test2", "test2@mc.org")
    project.id = "test2"
    insert(project.__dict__, "projects", conn)
    ddir = DataDir("test2", "test2@mc.org", "")
    ddir.id = "test2"
    ddir.project = project.id
    insert(ddir.__dict__, "datadirs", conn)
    project2datadir = {
        "project_id": project.id,
        "datadir_id": ddir.id
    }
    insert(project2datadir, "project2datadir", conn)
    uaccess = Access("test2", "test2", "test2@mc.org")
    insert(uaccess.__dict__, "access", conn)

    project = Project("test3", "test3@mc.org")
    project.id = "test3"
    insert(project.__dict__, "projects", conn)
    ddir = DataDir("test3", "test3@mc.org", "")
    ddir.id = "test3"
    ddir.project = project.id
    insert(ddir.__dict__, "datadirs", conn)
    project2datadir = {
        "project_id": project.id,
        "datadir_id": ddir.id
    }
    insert(project2datadir, "project2datadir", conn)
    uaccess = Access("test3", "test3", "test3@mc.org")
    insert(uaccess.__dict__, "access", conn)

    uaccess = Access("test3", "test3", "test@mc.org")
    insert(uaccess.__dict__, "access", conn)

    print "Done..."


def create_db():
    print "Creating mctestdb..."
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
