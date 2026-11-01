// fsshim.js — an in-memory filesystem for goloco's js/wasm build.
//
// Go's js/wasm runtime routes os.Open / ReadDir / Stat / Read through the
// Node-style globalThis.fs object (src/syscall/fs_js.go). Browsers have no such
// object, so wasm_exec.js installs stubs that fail with ENOSYS. This file
// installs a working one, backed by files unpacked from the data pack (see
// cmd/webpack), so goloco's loaders run unchanged: the game sees
//   /w/goloco            its working directory (data/, assets/)
//   /w/locomotion/...    the game data (Data/, ObjData/, Scenarios/)
// exactly as ../locomotion/Data resolves from the repo natively.
//
// Writes (progress file, screenshots) go to memory and are lost on reload.
// Every missing path is recorded in GolocoFS.misses so the page can show what
// the pack lacks.
"use strict";

(() => {
	const S_IFDIR = 0o040000, S_IFREG = 0o100000;
	const O_WRONLY = 1, O_RDWR = 2, O_CREAT = 64, O_EXCL = 128, O_TRUNC = 512, O_APPEND = 1024, O_DIRECTORY = 65536;

	const files = new Map();   // absolute path -> Uint8Array
	const dirs = new Map();    // absolute path -> Set of entry names
	const fds = new Map();     // fd -> {path, pos, dir, append}
	const misses = new Set();
	let nextFd = 10;
	let cwd = "/w/goloco";
	let out = { 1: "", 2: "" };
	const decoder = new TextDecoder("utf-8");

	const norm = (p) => {
		if (!p.startsWith("/")) p = cwd + "/" + p;
		const parts = [];
		for (const s of p.split("/")) {
			if (s === "" || s === ".") continue;
			if (s === "..") parts.pop(); else parts.push(s);
		}
		return "/" + parts.join("/");
	};
	const parent = (p) => p.slice(0, p.lastIndexOf("/")) || "/";
	const base = (p) => p.slice(p.lastIndexOf("/") + 1);

	const mkdirp = (p) => {
		while (!dirs.has(p)) {
			dirs.set(p, new Set());
			if (p === "/") break;
			dirs.get(parent(p)) || mkdirp(parent(p));
			dirs.get(parent(p)).add(base(p));
			p = parent(p);
		}
	};
	mkdirp("/w/goloco");
	mkdirp("/w/locomotion");

	const addFile = (path, data) => {
		path = norm(path);
		mkdirp(parent(path));
		dirs.get(parent(path)).add(base(path));
		files.set(path, data);
	};

	const err = (code, path) => {
		if (code === "ENOENT" && path) misses.add(path);
		const e = new Error(code + (path ? ": " + path : ""));
		e.code = code;
		return e;
	};

	const statOf = (path) => {
		const isDir = dirs.has(path);
		const f = files.get(path);
		if (!isDir && !f) return null;
		return {
			dev: 0, ino: 0, mode: isDir ? (S_IFDIR | 0o755) : (S_IFREG | 0o644), nlink: 1,
			uid: 0, gid: 0, rdev: 0, size: isDir ? 0 : f.length, blksize: 4096,
			blocks: isDir ? 0 : Math.ceil(f.length / 512), atimeMs: 0, mtimeMs: 0, ctimeMs: 0,
			isDirectory() { return isDir; },
		};
	};

	const console_out = (fd, s) => {
		out[fd] += s;
		const nl = out[fd].lastIndexOf("\n");
		if (nl !== -1) {
			const lines = out[fd].substring(0, nl);
			out[fd] = out[fd].substring(nl + 1);
			(fd === 2 ? console.warn : console.log)(lines);
			if (globalThis.GolocoFS.onLog) globalThis.GolocoFS.onLog(lines);
		}
	};

	const writeAt = (fd, buf, offset, length, position) => {
		const f = fds.get(fd);
		const old = files.get(f.path) || new Uint8Array(0);
		const at = f.append ? old.length : (position ?? f.pos);
		const end = at + length;
		let data = old;
		if (end > old.length) { data = new Uint8Array(end); data.set(old); }
		data.set(buf.subarray(offset, offset + length), at);
		files.set(f.path, data);
		if (position == null) f.pos = end;
		return length;
	};

	const fs = {
		constants: { O_WRONLY, O_RDWR, O_CREAT, O_TRUNC, O_APPEND, O_EXCL, O_DIRECTORY },
		writeSync(fd, buf) {
			if (fd === 1 || fd === 2) { console_out(fd, decoder.decode(buf)); return buf.length; }
			return writeAt(fd, buf, 0, buf.length, null);
		},
		write(fd, buf, offset, length, position, cb) {
			if (fd === 1 || fd === 2) { console_out(fd, decoder.decode(buf.subarray(offset, offset + length))); cb(null, length); return; }
			if (!fds.has(fd)) { cb(err("EBADF")); return; }
			cb(null, writeAt(fd, buf, offset, length, position));
		},
		open(path, flags, mode, cb) {
			path = norm(path);
			const exists = files.has(path) || dirs.has(path);
			if (!exists && !(flags & O_CREAT)) { cb(err("ENOENT", path)); return; }
			if (exists && (flags & O_CREAT) && (flags & O_EXCL)) { cb(err("EEXIST", path)); return; }
			if ((flags & O_DIRECTORY) && !dirs.has(path)) { cb(err(exists ? "ENOTDIR" : "ENOENT", path)); return; }
			if (!exists) {
				if (!dirs.has(parent(path))) { cb(err("ENOENT", path)); return; }
				addFile(path, new Uint8Array(0));
			} else if ((flags & O_TRUNC) && files.has(path)) {
				files.set(path, new Uint8Array(0));
			}
			const fd = nextFd++;
			fds.set(fd, { path, pos: 0, dir: dirs.has(path), append: !!(flags & O_APPEND) });
			cb(null, fd);
		},
		read(fd, buffer, offset, length, position, cb) {
			const f = fds.get(fd);
			if (!f) { cb(err("EBADF")); return; }
			if (f.dir) { cb(err("EISDIR", f.path)); return; }
			const data = files.get(f.path);
			const at = position ?? f.pos;
			const n = Math.max(0, Math.min(length, data.length - at));
			buffer.set(data.subarray(at, at + n), offset);
			if (position == null) f.pos += n;
			cb(null, n);
		},
		close(fd, cb) { fds.delete(fd); cb(null); },
		fstat(fd, cb) {
			const f = fds.get(fd);
			if (!f) { cb(err("EBADF")); return; }
			cb(null, statOf(f.path));
		},
		stat(path, cb) { path = norm(path); const s = statOf(path); s ? cb(null, s) : cb(err("ENOENT", path)); },
		lstat(path, cb) { this.stat(path, cb); },
		readdir(path, cb) {
			path = norm(path);
			const d = dirs.get(path);
			d ? cb(null, [...d].sort()) : cb(err("ENOENT", path));
		},
		mkdir(path, perm, cb) {
			path = norm(path);
			if (dirs.has(path) || files.has(path)) { cb(err("EEXIST", path)); return; }
			if (!dirs.has(parent(path))) { cb(err("ENOENT", path)); return; }
			mkdirp(path); cb(null);
		},
		unlink(path, cb) {
			path = norm(path);
			if (!files.delete(path)) { cb(err("ENOENT", path)); return; }
			dirs.get(parent(path)).delete(base(path)); cb(null);
		},
		rename(from, to, cb) {
			from = norm(from); to = norm(to);
			const data = files.get(from);
			if (!data) { cb(err("ENOENT", from)); return; }
			files.delete(from); dirs.get(parent(from)).delete(base(from));
			addFile(to, data); cb(null);
		},
		rmdir(path, cb) { cb(err("ENOSYS")); },
		fsync(fd, cb) { cb(null); },
		chmod(p, m, cb) { cb(null); }, fchmod(fd, m, cb) { cb(null); },
		chown(p, u, g, cb) { cb(null); }, fchown(fd, u, g, cb) { cb(null); }, lchown(p, u, g, cb) { cb(null); },
		utimes(p, a, m, cb) { cb(null); },
		ftruncate(fd, len, cb) { const f = fds.get(fd); files.set(f.path, (files.get(f.path) || new Uint8Array(0)).slice(0, len)); cb(null); },
		truncate(p, len, cb) { p = norm(p); files.set(p, (files.get(p) || new Uint8Array(0)).slice(0, len)); cb(null); },
		readlink(p, cb) { cb(err("EINVAL")); }, link(p, l, cb) { cb(err("ENOSYS")); }, symlink(p, l, cb) { cb(err("ENOSYS")); },
	};

	globalThis.fs = fs;
	globalThis.process = {
		getuid() { return -1; }, getgid() { return -1; }, geteuid() { return -1; }, getegid() { return -1; },
		getgroups() { return []; }, pid: 1, ppid: 0, umask() { return 0o22; },
		cwd() { return cwd; },
		chdir(p) { p = norm(p); if (!dirs.has(p)) throw err("ENOENT", p); cwd = p; },
	};
	globalThis.path = { resolve(...segs) { return norm(segs.join("/")); } };
	globalThis.GolocoFS = { addFile, misses, files, dirs, onLog: null };
})();
