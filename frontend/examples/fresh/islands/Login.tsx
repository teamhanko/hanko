import { HANKO_API_ENDPOINT } from "../config.ts";

const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.5.5-beta';

  register('${HANKO_API_ENDPOINT}', { shadow: true });
  document.addEventListener('hankoAuthSuccess', (event) => {
    document.location.href = '/todo';
  });
`;

export default function Login() {
  return (
    <div>
      <hanko-auth api={HANKO_API_ENDPOINT}></hanko-auth>
      <script type="module">
        {code}
      </script>
    </div>
  );
}
