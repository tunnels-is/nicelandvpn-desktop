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
