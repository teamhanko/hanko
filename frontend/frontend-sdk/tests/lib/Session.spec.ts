import { Session } from "../../src/lib/Session";
import { SessionDetail } from "../../src";

describe("Session", () => {
  let session: Session;

  beforeEach(() => {
    session = new Session();
  });

  describe("get", () => {
    it("should return session details if valid", () => {
      // Prepare
      const expectedDetails = {
        userID: "12345",
        expirationSeconds: 3600,
        jwt: "some.jwt.token",
      };

      // Mock dependencies
      jest.spyOn(session["_sessionState"], "read").mockImplementation();
      jest
        .spyOn(session["_sessionState"], "getUserID")
        .mockReturnValue(expectedDetails.userID);
      jest
        .spyOn(session["_sessionState"], "getExpirationSeconds")
        .mockReturnValue(expectedDetails.expirationSeconds);
      jest
        .spyOn(session["_cookie"], "getAuthCookie")
        .mockReturnValue(expectedDetails.jwt);

      // Execute
      const result = session.get();

      // Verify
      expect(result).toEqual(expectedDetails);
    });

    it("should return null if session details are invalid", () => {
      // Prepare
      const invalidDetails: SessionDetail = {
        userID: "",
        expirationSeconds: 0,
        jwt: null,
      };

      // Mock dependencies
      jest.spyOn(session["_sessionState"], "read").mockImplementation();
      jest
        .spyOn(session["_sessionState"], "getUserID")
        .mockReturnValue(invalidDetails.userID);
      jest
        .spyOn(session["_sessionState"], "getExpirationSeconds")
        .mockReturnValue(invalidDetails.expirationSeconds);
      jest
        .spyOn(session["_cookie"], "getAuthCookie")
        .mockReturnValue(invalidDetails.jwt);

      // Execute
      const result = session.get();

      // Verify
      expect(result).toBeNull();
    });
  });

  describe("isValid", () => {
    it("should return true if the user is logged in", () => {
      // Prepare
      const loggedInDetails = {
        userID: "12345",
        expirationSeconds: 3600,
        jwt: "some.jwt.token",
      };

      // Mock dependencies
      jest.spyOn(session, "_get").mockReturnValue(loggedInDetails);

      // Execute
      const result = session.isValid();

      // Verify
      expect(result).toBe(true);
    });

    it("should return false if the user is not logged in", () => {
      // Prepare
      const notLoggedInDetails: SessionDetail = {
        userID: "",
        expirationSeconds: 0,
        jwt: null,
      };

      // Mock dependencies
      jest.spyOn(session, "_get").mockReturnValue(notLoggedInDetails);

      // Execute
      const result = session.isValid();

      // Verify
      expect(result).toBe(false);
    });
  });
});
