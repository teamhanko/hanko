import { HANKO_API_URL } from "../config.ts";

const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@1.0.0';

  register('${HANKO_API_URL}', { shadow: true });
  document.addEventListener('hankoAuthSuccess', (event) => {
    document.location.href = '/todo';
  });
`;

export default function Login() {
  return (
    <div>
      <hanko-auth api={HANKO_API_URL}></hanko-auth>
      <script type="module">
        {code}
      </script>
    </div>
  );
}
