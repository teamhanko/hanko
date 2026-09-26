import type {
  FlowAction,
  LoginInitActions,
  ProfileInitActions,
  ContinueWithLoginIdentifierInputs,
  PatchMetadataInputs,
} from "../../../src";

describe("public Flow API types", () => {
  it("exports action response types from the SDK entrypoint", () => {
    const loginAction: FlowAction<ContinueWithLoginIdentifierInputs> = {
      action: "continue_with_login_identifier",
      href: "/login?action=continue_with_login_identifier",
      inputs: {
        identifier: {
          name: "identifier",
          type: "string",
        },
      },
      description: "Continue with a login identifier",
    };
    const profileAction: FlowAction<PatchMetadataInputs> = {
      action: "patch_metadata",
      href: "/profile?action=patch_metadata",
      inputs: {
        patch_metadata: {
          name: "patch_metadata",
          type: "object",
        },
      },
      description: "Update profile metadata",
    };
    const loginActions: LoginInitActions = {
      continue_with_login_identifier: loginAction,
    };
    const profileActions: Pick<ProfileInitActions, "patch_metadata"> = {
      patch_metadata: profileAction,
    };

    expect(loginActions.continue_with_login_identifier).toBe(loginAction);
    expect(profileActions.patch_metadata).toBe(profileAction);
  });
});
