const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.2.2-alpha';

  register({ shadow: true });
  document.addEventListener('hankoAuthSuccess', (event) => {
    document.location.href = '/todo';
  });
`;

const endpoint = "HANKO_API_ENDPOINT";

export default function Login() {
  return (
    <div>
      <hanko-auth api={endpoint}></hanko-auth>
      <script type="module">
        {code}
      </script>
    </div>
  );
}
