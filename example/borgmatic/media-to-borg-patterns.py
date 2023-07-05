#!/usr/bin/env python3
import argparse
import os
import pathlib
import sys

def main():
	cli = argparse.ArgumentParser(
		prog="media-to-borg-patterns",
		description="""Generate Borg patterns to backup media files belonging to
		this instance. You can pass the output to Borg or Borgmatic as a patterns file.
		For example: gotosocial admin media list-local | media-to-borg-patterns
		<storage-local-base-path>. You can pass a second argument, the destination file, to
		write the patterns in. If it's ommitted the patterns will be emitted on stdout
		instead and you can redirect the output to a file yourself.
		""",
		epilog="Be gay, do backups. Trans rights!"
	)
	cli.add_argument("storageroot", type=pathlib.Path, help="same value as storage-local-base-path in your GoToSocial configuration")
	cli.add_argument("destination", nargs="?", type=pathlib.Path, help="file to write patterns to, or stdout if ommitted")
	args = cli.parse_args()

	output = open(args.destination, 'w') if args.destination else sys.stdout
	# Start recursing from the storage root, including the storage root itself
	output.write("R "+str(args.storageroot)+"\n")

	prefixes=set()

	for line in sys.stdin:
		# Skip any log lines
		if "msg=" in line:
			continue
		# Reduce the path to the storage path plus the account ID. By
		# doing this we can emit path-prefix patterns, one for each account,
		# instead of a path-file pattern for each file.
		prefixes.add(os.path.join(*line.split("/")[:-3]))

	for prefix in prefixes:
		# Add a path-prefix, pp:, for each path we want to include.
		output.write("+ pp:"+prefix+"\n")

	# Exclude every file and directory under the storage root. This excludes
	# everything that wasn't matched by any of our prior patterns. This turns
	# the emitted patterns into an "include only" list.
	output.write("- "+os.path.join(args.storageroot, "*")+"\n")

	if output is not sys.stdout:
		output.close()

if __name__ == "__main__":
	main()
