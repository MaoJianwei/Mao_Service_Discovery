set -x

cd ./webui/
npm install
npm run build
cd ./dist/

rm -rf ../../resource/static/css/
rm -rf ../../resource/static/js/
mv favicon.ico ../../resource/static/
mv index.html ../../resource/html/
mv ./css/ ../../resource/static/
mv ./js/ ../../resource/static/

cd ../../

set +x