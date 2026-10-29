pragma Singleton
pragma ComponentBehavior: Bound

import QtQuick
import Quickshell

Singleton {
    property var transport: DuxFilesService

    function status(path, callback) {
        transport.call("git.status", {
            "path": path
        }, callback);
    }

    function stage(path, paths, callback) {
        _mutate("git.stage", path, paths, callback);
    }

    function unstage(path, paths, callback) {
        _mutate("git.unstage", path, paths, callback);
    }

    function stageAll(path, callback) {
        _mutate("git.stageAll", path, [], callback);
    }

    function unstageAll(path, callback) {
        _mutate("git.unstageAll", path, [], callback);
    }

    function entriesByPath(result) {
        const mapped = {};
        for (const entry of result?.entries ?? [])
            mapped[entry.path] = entry;
        return mapped;
    }

    function _mutate(method, path, paths, callback) {
        const params = {
            "path": path
        };
        if (paths.length > 0)
            params.paths = paths;
        transport.call(method, params, callback);
    }
}
