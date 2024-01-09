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


def is_suid(path: str) -> bool:
    """
    Checks if a file has the suid bit set.

    Parameters:
    path (str): The file to check
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
    sha1 = hashlib.sha1()
    with open(file, "rb") as f:
        data = f.read()
        sha1.update(data)
    return sha1.hexdigest()


def main(path: str, filelist: str, fsguard_binary: str):
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

                binaries.append(
                    "{} #FSG# {} #FSG# {}".format(
                        filepath, calc_checksum(filepath), "true" if suid else "false"
                    )
                )
    if not os.path.exists(filelist):
        file = open(filelist, "x")
        file.close()
    with open(filelist, "a") as file:
        file.write("\n".join(binaries) + "\n")


if __name__ == "__main__":
    main(path=sys.argv[1], filelist=sys.argv[2], fsguard_binary=sys.argv[3])
