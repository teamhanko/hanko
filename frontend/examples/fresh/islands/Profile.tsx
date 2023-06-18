const code = `
  import { register } from 'https://esm.sh/@teamhanko/hanko-elements@0.2.2-alpha';
  register({ shadow: true });
`;

const endpoint = "https://4c3bfaa1-a9b2-47ef-9c4f-ac01ee6d2e03.hanko.io";

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
