import os
import subprocess

basis_folder = "/mnt/data/results/"

files = []
i = 0
merge_procs = []

for root, dirs, s in os.walk(basis_folder):
    for filename in s:
        if filename.endswith(".lnk") or filename.endswith(".fixed"):
            pass
        else:
            files.append(f"{os.path.join(root, filename)}")
    if len(files) > 20:
        fix_procs = []
        for file in files:
            cmd = f"pcapfix '{file}' -k -o '{file}.fixed' && mv '{file}.fixed' '{file}'"
            fix_procs.append(subprocess.Popen(cmd, shell=True))
        for procs in fix_procs:
            try:
                procs.wait()
            except Exception as e:
                print(f"Error: {e}")
        file_names = []
        for file in files:
            file_names.append(f"'{file}'")
        if i > 1:
            cmd = f"mergecap -a -V -w merge{i}.pcap {' '.join(file_names)}"
            process = subprocess.Popen(cmd, shell=True)
            merge_procs.append(process)
        i += 1
        files = []

for process in merge_procs:
    process.wait()
