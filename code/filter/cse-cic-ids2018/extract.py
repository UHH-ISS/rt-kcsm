import os, zipfile, pyunpack
basis_folder =  os.getcwd()
destination_folder = "/mnt/data/results"

for root, dirs, files in os.walk(basis_folder):
    for filename in files:
        if not "pcap" in filename:
            continue
        if filename.endswith(".rar"):
            print('RAR:'+os.path.join(root,filename))
        elif filename.endswith(".zip"):
            print('ZIP:'+os.path.join(root,filename))
        name = os.path.splitext(os.path.basename(filename))[0]
        if filename.endswith(".rar") or filename.endswith(".zip"):
            try:
                dir_name = os.path.basename(root)
                arch = pyunpack.Archive(os.path.join(root,filename))
                dst = os.path.join(destination_folder, dir_name)
                os.makedirs(dst, exist_ok=True)
                arch.extractall(directory=dst)
            except Exception as e:
                print("ERROR: BAD ARCHIVE "+os.path.join(root,filename))
                print(e)
                try:
                    pass
                except OSError as e: # this would be "except OSError, e:" before Python 2.6
                    if e.errno != errno.ENOENT: # errno.ENOENT = no such file or directory
                        raise # re-raise exception if a different error occured   
