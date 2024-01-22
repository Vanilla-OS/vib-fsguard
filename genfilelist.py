# genfilelist.py
# Copyright 2024 axtlos <axtlos@getcryst.al>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, version 3 of the License only.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
#
# SPDX-License-Identifier: GPL-3.0-only

import os
import stat
import hashlib
import sys
import datetime


def log(message: str):
    """
    Logs a message with a timestamp and adds it to the global log list.

    Parameters:
    message (str): The message to log.
    """
    timestamp = datetime.datetime.now().strftime("[%Y-%m-%d %H:%M:%S]")
    log_list.append(f"{timestamp} {message}")
    print(message)


def print_help():
    """
    Prints the help message for using the script.
    """
    log(
        "Usage: python genfilelist.py <path> <filelist> <fsguard_binary> [--verbose] [--log-file <logfile>]"
    )


def is_suid(path: str) -> bool:
    """
    Checks if a file has the suid bit set.

    Parameters:
    path (str): The file to check

    Returns:
    bool: True if the suid bit is set, False otherwise.
    """
    binary = os.stat(path)
    if binary.st_mode & stat.S_ISUID == 2048:
        return True
    return False


def get_symlink(path: str) -> str:
    """
    Checks if a file is a symlink and returns the path it points to if it is.

    Parameters:
    path (str): The path to check

    Returns:
    str: The path the symlink points to. None if the file is not a symlink
    """
    if os.path.islink(path):
        return os.readlink(path)
    else:
        return ""


def calc_checksum(file: str) -> str:
    """
    Calculates the sha1sum of a given file

    Parameters:
    file (str): The file to calculate the checksum from

    Returns:
    str: The calculated checksum
    """
    data = None
    with open(file, "rb") as f:
        data = f.read()
    hash = hashlib.sha1(data).hexdigest()

    log(f"Checksum: {hash}")

    return hash


def main(
    path: str,
    filelist: str,
    fsguard_binary: str,
    verbose: bool = False,
    log_file: str = None,
):
    """
    Generates a file list with checksums and suid information for files in a given path.

    Parameters:
    path (str): The path to scan for files.
    filelist (str): The file to store the generated file list.
    fsguard_binary (str): The path to the fsguard binary.
    verbose (bool, optional): Enable verbose mode. Defaults to False.
    log_file (str, optional): Specify a log file to save verbose output. Defaults to None.
    """
    binaries: list[str] = []
    for dirpath, _, filenames in os.walk(path):
        for file in filenames:
            if not fsguard_binary.strip() in dirpath + "/" + file:
                filepath = dirpath + "/" + file
                filepath = os.path.abspath(
                    dirpath + "/" + get_symlink(filepath)
                    if get_symlink(filepath) != ""
                    else filepath
                )
                if not os.path.isfile(filepath):
                    filepath = os.path.abspath(get_symlink(dirpath + "/" + file))
                suid = is_suid(filepath)

                log(f"Processing: {filepath}")

                binaries.append(
                    "{} #FSG# {} #FSG# {}".format(
                        filepath, calc_checksum(filepath), "true" if suid else "false"
                    )
                )

                log(f"Processed: {filepath}\n")

    if not os.path.exists(filelist):
        file = open(filelist, "x")
        file.close()
    with open(filelist, "a") as file:
        file.write("\n".join(binaries) + "\n")

    if log_file:
        with open(log_file, "w") as log_file:
            log_file.write("\n".join(log_list) + "\n")


if __name__ == "__main__":
    if len(sys.argv) < 4 or "--help" in sys.argv:
        print_help()
        sys.exit(1)

    log_list = []
    path = sys.argv[1]
    filelist = sys.argv[2]
    fsguard_binary = sys.argv[3]
    verbose = "--verbose" in sys.argv
    log_file = None
    if "--log-file" in sys.argv:
        log_file_index = sys.argv.index("--log-file")
        log_file = (
            sys.argv[log_file_index + 1] if log_file_index + 1 < len(sys.argv) else None
        )
    main(
        path=path,
        filelist=filelist,
        fsguard_binary=fsguard_binary,
        verbose=verbose,
        log_file=log_file,
    )
