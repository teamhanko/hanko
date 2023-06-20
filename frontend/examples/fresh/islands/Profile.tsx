import { HANKO_API_ENDPOINT } from "../config.ts";

const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.5.5-beta';
  register('${HANKO_API_ENDPOINT}', { shadow: true });
`;

export default function Profile() {
  return (
    <div>
      <hanko-profile api={HANKO_API_ENDPOINT}></hanko-profile>
      <script type="module">
        {code}
      </script>
    </div>
  );
}
