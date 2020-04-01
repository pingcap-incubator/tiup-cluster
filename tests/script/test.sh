#!/bin/sh

cat run.sh | docker exec -i tiops-control bash -c "cat > /tmp/test.sh"
docker exec -i tiops-control bash -c "chmod +x /tmp/test.sh && /tmp/test.sh"
