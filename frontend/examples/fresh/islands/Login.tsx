const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.2.2-alpha';

  register({ shadow: true });
  document.addEventListener('hankoAuthSuccess', (event) => {
    document.location.href = '/todo';
  });
`;

const endpoint = "https://4c3bfaa1-a9b2-47ef-9c4f-ac01ee6d2e03.hanko.io";

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
