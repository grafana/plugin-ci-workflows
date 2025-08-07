set -e

if [ ! -d docs/sources ]; then echo "docs/sources not found. skipping build." && exit 0; fi

mkdir -p /hugo/content/docs/plugins/temp-name/v1.0.0
cp -r docs/sources /hugo/content/docs/plugins/temp-name/v1.0.0
make -C /hugo prod

echo "✅ Docs can be successfuly built"
