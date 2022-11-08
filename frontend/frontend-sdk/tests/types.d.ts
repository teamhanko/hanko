declare global {
  // eslint-disable-next-line no-unused-vars
  interface PublicKeyCredential {
    isExternalCTAP2SecurityKeySupported: () => Promise<boolean>;
    isConditionalMediationAvailable: () => Promise<boolean>;
  }
}

export {};
