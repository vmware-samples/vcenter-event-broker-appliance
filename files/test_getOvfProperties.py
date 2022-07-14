import pytest
import getOvfProperties
import json

# Note - these tests assume you are running them on a *nix machine with the `cat` command
# and in a shell that understands Heredocs (https://tldp.org/LDP/abs/html/here-docs.html)

# To run the tests ensure pytest is installed and execute `pytest test_getOvfProperty.py`
# from within this directory


def test_missingcommand(monkeypatch, capsys):
    monkeypatch.setattr(getOvfProperties, 'ovfenv_cmd', '/bin/madeupcommand -someparam')
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["missingvalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 1
    # assert captured.out == ""


def test_invalidxml(monkeypatch, capsys):
    monkeypatch.setattr(getOvfProperties, 'ovfenv_cmd', """cat << 'EOF'
<?xml version="1.0" encoding="UTF-8"
EOF""")
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["missingvalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 1
    # assert captured.out == ""


def test_missingvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["missingvalue"])
    captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    assert captured.err == "ovfProperty not found: 'guestinfo.missingvalue' \n"


def test_textvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["textvalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == 'TEXTVALUE="Hello World!"\n'
    assert get_veba_config('TEXTVALUE') == 'Hello World!'


def test_boolvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["boolvalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == 'BOOLVALUE="True"\n'
    assert get_veba_config('BOOLVALUE') == 'True'


def test_numbervalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["numbervalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == r"NUMBERVALUE='3.141'"
    assert get_veba_config('NUMBERVALUE') == '3.141'


def test_specialchars(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["specialchars"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == r"SPECIALCHARS='& < > ' \""
    assert get_veba_config('SPECIALCHARS') == '& < > \' "'


def test_doublequotedvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["doublequotedvalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == 'DOUBLEQUOTEDVALUE="Hello World"\n'
    assert get_veba_config('DOUBLEQUOTEDVALUE') == 'Hello World'


def test_singlequotedvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["singlequotedvalue"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == r"SINGLEQUOTEDVALUE='Hello World'"
    assert get_veba_config('SINGLEQUOTEDVALUE') == "Hello World"


def test_mismatchedquotesvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["mismatchedquotes"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == r"MISMATCHEDQUOTES=''\"'\"'Hello World\"'"
    assert get_veba_config('MISMATCHEDQUOTES') == '\'Hello World"'


def test_onequotevalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["onequote"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == r"ONEQUOTE='Hello World\"'"
    assert get_veba_config('ONEQUOTE') == 'Hello World"'


def test_passwordvalue(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["test_password"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == 'TEST_PASSWORD="\\"My&Quoted!Password\\""\n'
    assert get_veba_config('TEST_PASSWORD') == '"My&Quoted!Password"'


def test_shellspecials(capsys):
    with pytest.raises(SystemExit) as pytest_wrapped_e:
        getOvfProperties.main(["shellspecials"])
    # captured = capsys.readouterr()
    assert pytest_wrapped_e.type == SystemExit
    assert pytest_wrapped_e.value.code == 0
    # assert captured.out == r"SHELLSPECIALS='$`\"\!'"
    assert get_veba_config('SHELLSPECIALS') == r'$`"\!'


def get_veba_config(param):
    with open(getOvfProperties.veba_config_file) as fp:
        veba_config = json.load(fp)

    return veba_config[param]

# Patch the ovfenv_cmd global to return a dumy ovf env XML
@pytest.fixture(autouse=True)
def ovfenv(monkeypatch):
    monkeypatch.setattr(getOvfProperties, 'veba_config_file', './test_veba_config.json')
    monkeypatch.setattr(getOvfProperties, 'veba_env_file', './test_veba_env')
    monkeypatch.setattr(getOvfProperties, 'ovfenv_cmd', r"""cat << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<Environment
    xmlns="http://schemas.dmtf.org/ovf/environment/1"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xmlns:oe="http://schemas.dmtf.org/ovf/environment/1"
    xmlns:ve="http://www.vmware.com/schema/ovfenv"
    oe:id=""
    ve:vCenterId="vm-1137482">
<PlatformSection>
    <Kind>VMware ESXi</Kind>
    <Version>7.0.3</Version>
    <Vendor>VMware, Inc.</Vendor>
    <Locale>en_US</Locale>
</PlatformSection>
<PropertySection>
        <Property oe:key="guestinfo.textvalue" oe:value="Hello World!"/>
        <Property oe:key="guestinfo.boolvalue" oe:value="True"/>
        <Property oe:key="guestinfo.numbervalue" oe:value="3.141"/>
        <Property oe:key="guestinfo.specialchars" oe:value="&amp; &lt; &gt; &apos; &quot;"/>
        <Property oe:key="guestinfo.doublequotedvalue" oe:value="&quot;Hello World&quot;"/>
        <Property oe:key="guestinfo.singlequotedvalue" oe:value="&apos;Hello World&apos;"/>
        <Property oe:key="guestinfo.mismatchedquotes" oe:value="&apos;Hello World&quot;"/>
        <Property oe:key="guestinfo.onequote" oe:value="Hello World&quot;"/>
        <Property oe:key="guestinfo.test_password" oe:value="&quot;My&amp;Quoted!Password&quot;"/>
        <Property oe:key="guestinfo.shellspecials" oe:value="$`&quot;\!"/>
</PropertySection>
<ve:EthernetAdapterSection>
    <ve:Adapter ve:mac="00:50:56:ba:30:35" ve:network="management-101" ve:unitNumber="7"/>
</ve:EthernetAdapterSection>
</Environment>
EOF
""")
