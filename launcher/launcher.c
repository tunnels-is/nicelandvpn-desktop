#include <stdio.h>
#include <stdlib.h>
#include <Security/Authorization.h>
#import <string.h>
#import <unistd.h>
#import <CoreFoundation/CoreFoundation.h>

int main(int argc, char *argv[]) {

  CFURLRef url = CFBundleCopyExecutableURL(CFBundleGetMainBundle());
  char path[PATH_MAX];
  CFURLGetFileSystemRepresentation(url, true, (UInt8*)path, sizeof(path));
  CFRelease(url);
  int size = strlen(path);
  path[size-9] = '\0';
  char fullPath[PATH_MAX+12];
  strcpy(fullPath,path);
  strcat(fullPath, "/NicelandVPN");
  // printf("DIR:%s\n", fullPath);

  const char *authRightName = "system.privilege.admin";
    AuthorizationItem right = {authRightName, 0, NULL, 0};
    AuthorizationRights rights = {1, &right};
    AuthorizationFlags flags = kAuthorizationFlagExtendRights | kAuthorizationFlagInteractionAllowed | kAuthorizationFlagPreAuthorize;

    AuthorizationRef authRef = NULL;
    OSStatus status = AuthorizationCreate(NULL, kAuthorizationEmptyEnvironment, kAuthorizationFlagDefaults, &authRef);
    if (status != errAuthorizationSuccess) {
        // printf("Error creating authorization reference: %d\n", status);
        return 1;
    }

    status = AuthorizationCopyRights(authRef, &rights, NULL, flags, NULL);
    if (status != errAuthorizationSuccess) {
        // printf("Error acquiring rights: %d\n", status);
        AuthorizationFree(authRef, kAuthorizationFlagDestroyRights);
        return 1;
    }

    char *arguments[] = { NULL}; 
    status = AuthorizationExecuteWithPrivileges(authRef, fullPath, kAuthorizationFlagDefaults, arguments, NULL);
    if (status != errAuthorizationSuccess) {
        // printf("Error executing program: %d\n", status);
        AuthorizationFree(authRef, kAuthorizationFlagDestroyRights);
        return 1;
    }
    // printf("STATUS %d\n", status);

    AuthorizationFree(authRef, kAuthorizationFlagDestroyRights);

    return 0;
}