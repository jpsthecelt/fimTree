# fimTree
Enable File Integrity Monitoring on directory trees

The Need
--------
So, I've had multiple customers ask for a File-Integrity-Monitoring service in
conjunction with BigFix, which we don't really provide.

In the sister-repository, I've created some fixlets that do a poor but perhaps
acceptable job of calculating and validating checksums of indicated files
and alerting the administrator when differences emerge.  This is based on the
client-relevance which can calculate a file's SHA1 checksum using the relevance:
		 sha1 of file "blah"

One of the potential 'downsides' of this is that it's 'expensive', in real-time,
and the reporting/resetting is problematic.

What's needed is an easy way to calculate checksums of files/trees, go back and
revalidate them, report on differences, and finally, restore them to their
original configuration (based on a canonical or 'prototype' image).

Thinking about The Solution
---------------------------
I figured I'd 'tackle' the problem by finding a 'cheap' way of calculating SHA1s
of a directory-tree, store it somewhere (a DB perhaps), and then perform a
comparison/update process.

One thought, was the implementation-language to choose.  I had a couple of
candidates: Scala and Go. A couple of features I liked were that each had a
'compiler', were rumored to be relatively fast, had good sets of libraries,
were not limited computationally, and would run on a multitude of platforms.

Go:
---
I liked the fact that I could produce free-standing binaries for a variety of
platforms, and despite the lack of a source-line debugger, implementation
seemed pretty easy.

All the primitives were there, so my first-pass used a single thread, worked
it's way down a directory-tree, calculating all checksums.  The result was
ts.go. You can see some screenshots I've attached wherein I used the
cygwin commandline like:
	     ts.exe .
and the powershell command
    	     ts.exe $pwd

Now, another thing I liked was that both languages could easily utilize multiple
processor-cores, and primitives like message-queue abstractions allowed easy
implementation.

My first implementation was 48 lines (including whitespace),
courtesy of a primitive called filepath.walk() which provided a
lexicographically-ordered 'tree' of filenames/directories.

I then gradually evolved the implementation up to ts5.go, creating two queues,
one to hold the list of the 'work to be done' (from the filepath.walk output),
and one to take each workers results (a formatted string including the checksum)
and output it; initially to the screen.

So ts5.go now utilizes all 8 of my Thinkpad's CPUs, and the Outputter task
can be modified to submit the results to the screen, a file, or a database.

After initiating alal the threads, I 'sit on' a Scanln, waiting for the
calculations to finish. Once complete, the operator presses ENTER and all
subsidiary-threads are sent 0-length-strings, signalling termination.

I'll continue modifying this and get to a more usable 'product'.

jpsthecelt-080315
