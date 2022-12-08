// `.client.ts` files are only supposed to run on the client. I could not find this
// documented on the remix docs but someone on the remix discord told me this is how it is
// supposed to work.
import { register } from '@teamhanko/hanko-elements/hanko-auth';

export { register as registerHankoAuth };
