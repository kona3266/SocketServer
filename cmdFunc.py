import os
import time
def cmdExcute(command):
    if command=="date":
        return time.ctime()
    elif command=="os":
        return os.name
    elif command=="ls":
        return os.listdir(os.getcwd()) 
    #listdir返回指定的文件夹包含的文件或文件夹的名字的列表。
    #os.getcwd()返回当前目录名称。
    else:
        return command