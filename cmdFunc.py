import subprocess
import sys
def cmdExcute(cmd):
    cmd = cmd.split()
    os_type = sys.platform
    complete_process = subprocess.run(cmd,shell=True,capture_output=True)
    if 'win' in os_type:
        encoding = 'ansi'
    else:
        encoding = 'utf-8'
    stdout = complete_process.stdout.decode(encoding)
    stderr = complete_process.stderr.decode(encoding)
    if complete_process.returncode:
        return '{0}: {1}'.format(complete_process.returncode,stderr)
    else:
        return 'Output: {0}'.format(stdout)