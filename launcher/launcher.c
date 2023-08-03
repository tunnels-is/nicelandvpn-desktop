#include <stdio.h>
#include <Security/Authorization.h>

int main() {
    // Replace with the path to the program you want to execute with elevated privileges
    const char *programPath = "./NicelandVPN";

    // Create the AuthorizationRef object
    AuthorizationRef authRef;
    OSStatus status = AuthorizationCreate(NULL, kAuthorizationEmptyEnvironment, kAuthorizationFlagDefaults, &authRef);
    if (status != errAuthorizationSuccess) {
        printf("Error creating authorization reference: %d\n", status);
        return 1;
    }

    // Set the right to run the command as the root user
    AuthorizationItem authItems = {kAuthorizationRightExecute, 0, NULL, 0};
    AuthorizationRights authRights = {1, &authItems};

    // Acquire the necessary privileges
    status = AuthorizationCopyRights(authRef, &authRights, kAuthorizationEmptyEnvironment, kAuthorizationFlagPreAuthorize, NULL);
    if (status != errAuthorizationSuccess) {
        printf("Error acquiring rights: %d\n", status);
        AuthorizationFree(authRef, kAuthorizationFlagDestroyRights);
        return 1;
    }

    // Execute the program with administrator privileges
    FILE *output = NULL;
    char *arguments[] = {NULL}; // Add program arguments if needed
    status = AuthorizationExecuteWithPrivileges(authRef, programPath, kAuthorizationFlagDefaults, arguments, &output);
    if (status != errAuthorizationSuccess) {
        printf("Error executing program: %d\n", status);
    }

    // Clean up resources
    if (output != NULL) fclose(output);
    AuthorizationFree(authRef, kAuthorizationFlagDestroyRights);

    return 0;
}