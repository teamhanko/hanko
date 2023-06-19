const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.2.2-alpha';
  register({ shadow: true });
`;

const endpoint = "HANKO_API_ENDPOINT";

export default function Profile() {
  return (
    <div>
      <hanko-profile api={endpoint}></hanko-profile>
      <script type="module">
        {code}
      </script>
    </div>
  );
}
