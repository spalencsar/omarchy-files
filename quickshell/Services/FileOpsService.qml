pragma Singleton
pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Common
import qs.Services

Singleton {
    id: root

    property var transport: DuxFilesService

    readonly property bool connected: transport.connected

    property var jobs: []
    property string lastDestination: ""
    property bool notificationsEnabled: true
    property var notificationSender: command => Quickshell.execDetached(command)

    readonly property int activeCount: jobs.filter(job => job.state !== "done" && job.state !== "failed").length

    property var elevation: ({
            "available": false,
            "route": "",
            "reason": ""
        })
    readonly property bool canElevate: elevation.available === true

    signal requestFailed(string message)
    signal jobFinished(var job)

    property var _byId: ({})
    property var _notifiedFinished: ({})

    readonly property DaemonTopic _jobs: DaemonTopic {
        transport: root.transport
        topic: "jobs"
        handler: data => root._onEvent(data)
        onConnected: {
            root.refresh();
            root.refreshElevation();
        }
    }

    function copy(paths, dest, scope) {
        _start("files.op.copy", {
            "paths": paths,
            "dest": dest,
            "scope": scope || ""
        });
    }

    function move(paths, dest, scope) {
        _start("files.op.move", {
            "paths": paths,
            "dest": dest,
            "scope": scope || ""
        });
    }

    function rename(path, name, scope) {
        _start("files.op.rename", {
            "path": path,
            "name": name,
            "scope": scope || ""
        });
    }

    function newFolder(dir, name, scope, unique) {
        _start("files.op.newFolder", {
            "path": dir,
            "name": name,
            "unique": unique === true,
            "scope": scope || ""
        });
    }

    function newFile(dir, name, template, scope, unique) {
        _start("files.op.newFile", {
            "path": dir,
            "name": name,
            "template": template || "",
            "unique": unique === true,
            "scope": scope || ""
        });
    }

    function link(paths, dest, scope) {
        _start("files.op.link", {
            "paths": paths,
            "dest": dest,
            "scope": scope || ""
        });
    }

    function trash(paths, scope) {
        _start("files.op.trash", {
            "paths": paths,
            "scope": scope || ""
        });
    }

    function restore(names, trashDir, scope) {
        _start("files.op.restore", {
            "names": names,
            "trashDir": trashDir,
            "scope": scope || ""
        });
    }

    function deletePermanently(paths) {
        _start("files.op.delete", {
            "paths": paths
        });
    }

    function emptyTrash() {
        _start("files.op.emptyTrash", null);
    }

    function compress(paths, dest, name, format, password, scope) {
        _start("files.op.compress", {
            "paths": paths,
            "dest": dest,
            "name": name,
            "format": format || "zip",
            "password": password || "",
            "scope": scope || ""
        });
    }

    function extract(paths, dest, scope) {
        _start("files.op.extract", {
            "paths": paths,
            "dest": dest,
            "scope": scope || ""
        });
    }

    function setPermissions(paths, mode, recursive) {
        _start("files.op.setPermissions", {
            "paths": paths,
            "mode": mode,
            "recursive": recursive === true
        });
    }

    function setOwner(paths, owner, group, recursive) {
        _start("files.op.setOwner", {
            "paths": paths,
            "owner": owner || "",
            "group": group || "",
            "recursive": recursive === true
        });
    }

    function respond(jobId, answerId, action, applyToAll) {
        _start("files.op.respond", {
            "jobId": jobId,
            "id": answerId,
            "action": action,
            "applyToAll": applyToAll === true
        });
    }

    function unlock(jobId, answerId, password) {
        _start("files.op.respond", {
            "jobId": jobId,
            "id": answerId,
            "action": "unlock",
            "password": password
        });
    }

    function cancel(jobId) {
        _start("files.op.cancel", {
            "jobId": jobId
        });
    }

    function undo(scope) {
        _start("files.op.undo", {
            "scope": scope || ""
        });
    }

    function forget(scope) {
        transport.sendRequest("files.op.forget", {
            "scope": scope || ""
        }, null);
    }

    function redo(scope) {
        _start("files.op.redo", {
            "scope": scope || ""
        });
    }

    function duplicate(paths, scope) {
        _start("files.op.duplicate", {
            "paths": paths,
            "scope": scope || ""
        });
    }

    function newFolderWith(paths, dir, name, scope, unique) {
        _start("files.op.newFolderWith", {
            "paths": paths,
            "path": dir,
            "name": name,
            "unique": unique === true,
            "scope": scope || ""
        });
    }

    function pasteImage(dir, name, scope) {
        _start("files.op.pasteImage", {
            "path": dir,
            "name": name,
            "scope": scope || ""
        });
    }

    function refresh() {
        transport.sendRequest("files.op.jobs", null, response => {
            if (response.error)
                return;
            _byId = ({});
            for (const job of (response.result || {}).jobs || [])
                _byId[job.id] = job;
            _rebuild();
        });
    }

    function refreshElevation() {
        transport.sendRequest("system.elevation", null, response => {
            if (response.error)
                return;
            const result = response.result || {};
            elevation = {
                "available": result.available === true,
                "route": result.route || "",
                "reason": result.reason || ""
            };
        });
    }

    function _start(method, params) {
        transport.sendRequest(method, params, response => {
            if (response.error) {
                root.requestFailed(response.error);
                return;
            }
            const job = (response.result || {}).job;
            if (job)
                root._store(job);
        });
    }

    function _onEvent(data) {
        switch (data.kind) {
        case "job":
            _onJob(data.job);
            return;
        case "progress":
            _onProgress(data.jobId, data.progress);
            return;
        }
    }

    function _onJob(job) {
        if (!job)
            return;
        const finished = job.state === "done" || job.state === "failed";
        _store(job);
        if (!finished)
            return;
        jobFinished(job);
        _notifyFinished(job);
        _historyTimeout.restart();
    }

    function _notifyFinished(job) {
        if (!notificationsEnabled || _notifiedFinished[job.id] === true || !_isTransfer(job.kind))
            return;
        _notifiedFinished[job.id] = true;

        const failed = job.state === "failed";
        const cancelled = job.cancelled === true;
        const title = failed
            ? I18n.tr("File operation failed", "desktop notification title for a failed transfer")
            : cancelled
            ? I18n.tr("File operation cancelled", "desktop notification title for a cancelled transfer")
            : _completionTitle(job.kind);
        const body = failed && (job.error || job.code)
            ? (job.error || job.code)
            : _completionBody(job);

        notificationSender([
            "omarchy", "notification", "send",
            "--app-name", "Omarchy Files",
            "-i", "io.github.spalencsar.OmarchyFiles",
            "-u", failed ? "critical" : "normal",
            "-t", "6000",
            title, body,
            "--exec", "omarchy-files", "show"
        ]);
    }

    function _isTransfer(kind) {
        return ["copy", "move", "compress", "extract", "duplicate", "link", "newFolderWith"].includes(kind);
    }

    function _completionTitle(kind) {
        switch (kind) {
        case "copy":
            return I18n.tr("Copy complete", "desktop notification title after copying files");
        case "move":
            return I18n.tr("Move complete", "desktop notification title after moving files");
        case "compress":
            return I18n.tr("Archive created", "desktop notification title after compressing files");
        case "extract":
            return I18n.tr("Extraction complete", "desktop notification title after extracting files");
        case "duplicate":
            return I18n.tr("Duplication complete", "desktop notification title after duplicating files");
        case "link":
            return I18n.tr("Links created", "desktop notification title after linking files");
        case "newFolderWith":
            return I18n.tr("Folder created", "desktop notification title after creating a folder from files");
        default:
            return I18n.tr("File operation complete", "desktop notification title after a transfer");
        }
    }

    function _completionBody(job) {
        const sources = job?.sources ?? [];
        const source = sources.length === 1
            ? _baseName(sources[0])
            : I18n.tr("%1 items", "number of items in a completed file operation").arg(sources.length);
        const destination = _baseName(job?.dest ?? "");
        if (source !== "" && destination !== "")
            return I18n.tr("%1 → %2", "desktop notification body: source and destination").arg(source).arg(destination);
        return source || destination || I18n.tr("The file operation finished successfully.", "desktop notification body after a transfer");
    }

    function _baseName(path) {
        const parts = String(path || "").split("/").filter(part => part !== "");
        return parts.length > 0 ? parts[parts.length - 1] : "";
    }

    readonly property Timer _historyTimeout: Timer {
        interval: 30000
        onTriggered: root._dropFinished()
    }

    function _dropFinished() {
        const live = ({});
        for (const id in _byId) {
            const job = _byId[id];
            if (job.state === "done" || job.state === "failed")
                continue;
            live[id] = job;
        }
        _byId = live;
        _rebuild();
    }

    function _onProgress(jobId, progress) {
        const job = _byId[jobId];
        if (!job)
            return;
        job.progress = progress;
        _rebuild();
    }

    function _store(job) {
        _byId[job.id] = job;
        _rebuild();
    }

    function _rebuild() {
        const list = [];
        for (const id in _byId)
            list.push(_byId[id]);
        list.sort((a, b) => a.startedMs - b.startedMs);
        jobs = list;
    }
}
