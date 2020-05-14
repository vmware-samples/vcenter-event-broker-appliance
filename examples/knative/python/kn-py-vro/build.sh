trap "exit" INT

RED='\033[0;31m'
GREEN='\033[0;32m'
ORANGE='\033[0;33m'
BLUE='\033[1;34m'
NC='\033[0m'

IMAGE="pilue/kn-py-vro"
PORT=8080

if [[ "$1" != "--no-tests" ]]; then
  pytest -v

  if [ $? -ne 0 ]; then
    echo -e "${RED}Tests failed - cancelling build${NC}"
    exit 1
  fi
  echo -e "${GREEN}All tests passed!${NC}"
else
  echo -e "${ORANGE}Skipping tests${NC}"
fi

echo -e "${BLUE}Building image ${IMAGE}...${NC}"

pack build --builder gcr.io/buildpacks/builder:v1 ${IMAGE}
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to run docker container - aborting${NC}"
  exit 2
fi
echo -e "${GREEN}Build completed - running locally with a test event...${NC}"
docker run --name "kn-py-vro" -d -e PORT=${PORT} -it --rm -p ${PORT}:${PORT} --env VROCONFIG_SECRET="$(cat kn-py-vro_secret.json)" ${IMAGE} > /dev/null
if [ $? -ne 0 ]; then
  echo -e "${RED}Failed to run docker container - aborting${NC}"
  exit 3
fi
sleep 6

res=$(curl -s -d@tests/testevent.json localhost:${PORT})

docker stop kn-py-vro > /dev/null 2>&1

echo "Result:"
echo $res | jq .

code=$(echo $res | jq .status)

if [[ "$code" == "202" ]]; then
  echo -e "${GREEN}Request to local function passed${NC}"
  exit 0
else
  echo -e "${RED}Request to local function failed${NC}"
  exit 4
fi