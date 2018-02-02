# RT vs FTP, SCP, RSync etc.

Are you really still deploying applications via FTP, SCP, rsync or any similar file transfer protocol? :anguished:

Let me tell you why it's a :poop: idea:

 - not all applications consist of files (you wouldn't deploy docker containers or Lambda functions like this)
 - consistency can be hard to achieve in some protocols (i.e. you may end up having half of updates on the server)
 - atomic deployments may be tricky to achieve without some extra symlink-swap logic
 - some of the mentioned procols can be insecure or difficult to secure reliably (yes, we have chroot, yes we can limit unix user capabilities, but it's not entirely enjoyble type of work...)
 - such deployment patterns imply we have a scalable and highly available file storage, which can be expensive and time consuming to maintain
