set -e

docs_source_directory="${1:-docs/sources}"

if [ ! -d "${docs_source_directory}" ]; then
  echo "${docs_source_directory} not found. skipping build." && exit 0
fi

mkdir -p /hugo/content/docs/plugins/temp-name/latest
cp -r "${docs_source_directory}"/. /hugo/content/docs/plugins/temp-name/latest/
make -C /hugo prod

echo "✅ Docs can be successfuly built"
