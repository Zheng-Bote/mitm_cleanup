import os
from datetime import datetime

date_str = datetime.now().strftime("%Y-%m-%d")
changelog_msg = """
### Fixed

- **SQL Schema Mismatch**: Fixed an issue where the cleanup queries for `job_audit_logs`, `admin_audit_logs`, `system_logs`, and `job_status_events` incorrectly referenced `created_at` instead of `ts`. (Fixes #2)
- **Incomplete Audit Logs**: By fixing the SQL schema mismatch, the cleanup job now successfully completes all deletions and properly sends the `ipc.SendAudit` messages for all tables to the scheduler. (Fixes #1)
"""

changelog_path = "CHANGELOG.md"
with open(changelog_path, "r") as f:
    cl_content = f.read()

new_entry = f"## [v0.8.3] - {date_str}\n{changelog_msg}\n"
insert_idx = cl_content.find("## [v")
cl_content = cl_content[:insert_idx] + new_entry + cl_content[insert_idx:]

with open(changelog_path, "w") as f:
    f.write(cl_content)
