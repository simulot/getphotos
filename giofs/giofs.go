package giofs

import "context"

/*
  Some documentation about gio command:

	https://manpages.ubuntu.com/manpages/jammy/en/man1/gio.1.html
	https://access.redhat.com/documentation/en-us/red_hat_enterprise_linux/8/html/using_the_desktop_environment_in_rhel_8/managing-storage-volumes-in-gnome_using-the-desktop-environment-in-rhel-8


*/
type FS struct {
	uri string
	ctx context.Context
}

// GIOFS create a fs.FS implementation using gio command on a linux system
func GIOFS(name string) *FS {
	return GIOFSWithContext(context.TODO(), name)
}

// GIOFS create a fs.FS implementation using gio command on a linux system and pass the given context
func GIOFSWithContext(ctx context.Context, name string) *FS {
	return &FS{
		ctx: ctx,
		uri: name,
	}

}
