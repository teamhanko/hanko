import { HANKO_API_URL } from "../config.ts";

const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.12.0';
  register('${HANKO_API_URL}', { shadow: true });
`;

export default function Profile() {
  return (
    <div>
      <hanko-profile api={HANKO_API_URL}></hanko-profile>
      <script type="module">
        {code}
      </script>
    </div>
  );
}
