#include <Security/Authorization.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

int requestElevatedPrivileges()
{
    const char *authRightName = "system.privilege.admin";
    AuthorizationItem right = {authRightName, 0, NULL, 0};
    AuthorizationRights rights = {1, &right};
    AuthorizationFlags flags = kAuthorizationFlagExtendRights | kAuthorizationFlagInteractionAllowed | kAuthorizationFlagPreAuthorize;

    AuthorizationRef authRef = NULL;
    OSStatus status = AuthorizationCreate(NULL, kAuthorizationEmptyEnvironment, kAuthorizationFlagDefaults, &authRef);
    if (status != errAuthorizationSuccess)
    {
        printf("STATUS ERROR: %d", status);
        return (int)status;
    }

    status = AuthorizationCopyRights(authRef, &rights, NULL, flags, NULL);
    printf("STATUS: %d", status);
    // AuthorizationFree(authRef, kAuthorizationFlagDefaults);
    return (int)status;
}

FILE *ExecuteStuff()
{

    FILE *fp;
    char path[1035];

    /* Open the command for reading. */
    fp = popen("networksetup", "r");
    if (fp == NULL)
    {
        printf("Failed to run command\n");
        exit(1);
    }
    printf("did run command!\n");

    /* Read the output a line at a time - output it. */
    // while (fgets(path, sizeof(path), fp) != NULL)
    // {
    //     printf("%s", path);
    // }

    // /* close */
    // pclose(fp);
    // (void)strncpy(dst, path, sizeof(path));

    return FILE;
}
