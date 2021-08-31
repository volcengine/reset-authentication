$LATEST_RESETAUTH="C:\Program Files\Reset Authentication\reset-authentication_latest.exe"
$RESETAUTH="C:\Program Files\Reset Authentication\reset-authentication.exe"

$HAS_LATEST=(Test-Path $LATEST_RESETAUTH)

if($HAS_LATEST -eq "True")
{
    Move-Item $LATEST_RESETAUTH $RESETAUTH -force
}

&$RESETAUTH

exit 1002
