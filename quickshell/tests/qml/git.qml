pragma ComponentBehavior: Bound

import QtQuick
import Quickshell
import qs.Common
import qs.Services

ShellRoot {
    id: fixture

    readonly property var log: Log.scoped("fixture")
    property int failures: 0
    property var sent: []

    function check(label, actual, expected) {
        const ok = JSON.stringify(actual) === JSON.stringify(expected);
        if (!ok) {
            failures++;
            log.error(`FAIL ${label}: got ${JSON.stringify(actual)} want ${JSON.stringify(expected)}`);
            return;
        }
        log.info(`ok ${label}`);
    }

    readonly property QtObject fakeTransport: QtObject {
        function call(method, params, callback) {
            fixture.sent = fixture.sent.concat([{ "method": method, "params": params }]);
            if (callback)
                callback({
                    "repository": true,
                    "root": "/repo",
                    "branch": "feature/git",
                    "entries": [{ "path": "/repo/a.txt", "code": "modified" }]
                });
        }
    }

    Component.onCompleted: {
        GitService.transport = fakeTransport;

        GitService.status("/repo", result => {
            check("status result", result.branch, "feature/git");
            check("entry mapping", GitService.entriesByPath(result)["/repo/a.txt"].code, "modified");
        });
        GitService.stage("/repo", ["/repo/a.txt"], null);
        GitService.unstage("/repo", ["/repo/a.txt"], null);
        GitService.stageAll("/repo", null);
        GitService.unstageAll("/repo", null);

        check("requests", sent, [
            { "method": "git.status", "params": { "path": "/repo" } },
            { "method": "git.stage", "params": { "path": "/repo", "paths": ["/repo/a.txt"] } },
            { "method": "git.unstage", "params": { "path": "/repo", "paths": ["/repo/a.txt"] } },
            { "method": "git.stageAll", "params": { "path": "/repo" } },
            { "method": "git.unstageAll", "params": { "path": "/repo" } }
        ]);

        if (failures === 0)
            log.info("FIXTURE PASS");
        Qt.callLater(Qt.quit);
    }
}
