const endpoint = "https://4c3bfaa1-a9b2-47ef-9c4f-ac01ee6d2e03.hanko.io";

const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.2.2-alpha';
  import { Hanko } from 'https://esm.sh/@teamhanko/hanko-frontend-sdk@0.5.4-beta';

  register({ shadow: true });
  window.addEventListener('logout', () => {
    const hanko = new Hanko('${endpoint}');
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

  return <button onClick={logout}>Logout</button>;
}
