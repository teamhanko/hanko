import { HANKO_API_ENDPOINT } from "../config.ts";

const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.2.2-alpha';
  import { Hanko } from 'https://esm.sh/@teamhanko/hanko-frontend-sdk@0.5.4-beta';

  register({ shadow: true });
  window.addEventListener('logout', () => {
    const hanko = new Hanko('${HANKO_API_ENDPOINT}');
    hanko.user.logout()
      .then(() => {
        window.location.href = '/';
      })
      .catch((error) => {
        alert(error);
      });
  });
`;


export default function Profile() {
  const logout = () => {
    window.dispatchEvent(new Event("logout"));
  };

  return (
    <>
      <button onClick={logout}>Logout</button>
      <script type="module">
        {code}
      </script>
    </>
  );
}
