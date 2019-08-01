package compiler

import (
	"github.com/hashload/boss/consts"
	"github.com/hashload/boss/utils"
	"io/ioutil"
	"os"
	"path/filepath"
)

func generateProject(paths []string, name string, savePath string) bool {
	generateDproj(savePath)
	return generateDpk(paths, name, savePath)
}

func generateDpk(paths []string, name string, savePath string) bool {
	var file = "package " + name + ";\n"
	file += "\n"
	file += "{$R *.res}\n"
	file += "{$IFDEF IMPLICITBUILDING This IFDEF should not be used by users}"
	file += "{$ALIGN 8}\n"
	file += "{$ASSERTIONS ON}\n"
	file += "{$BOOLEVAL OFF}\n"
	file += "{$DEBUGINFO OFF}\n"
	file += "{$EXTENDEDSYNTAX ON}\n"
	file += "{$IMPORTEDDATA ON}\n"
	file += "{$IOCHECKS ON}\n"
	file += "{$LOCALSYMBOLS ON}\n"
	file += "{$LONGSTRINGS ON}\n"
	file += "{$OPENSTRINGS ON}\n"
	file += "{$OPTIMIZATION OFF}\n"
	file += "{$OVERFLOWCHECKS OFF}\n"
	file += "{$RANGECHECKS OFF}\n"
	file += "{$REFERENCEINFO ON}\n"
	file += "{$SAFEDIVIDE OFF}\n"
	file += "{$STACKFRAMES ON}\n"
	file += "{$TYPEDADDRESS OFF}\n"
	file += "{$VARSTRINGCHECKS ON}\n"
	file += "{$WRITEABLECONST OFF}\n"
	file += "{$MINENUMSIZE 1}\n"
	file += "{$IMAGEBASE $400000}\n"
	file += "{$DEFINE DEBUG}\n"
	file += "{$ENDIF IMPLICITBUILDING}\n"
	file += "{$IMPLICITBUILD ON}\n\n\n"

	file += "contains\n"
	for key, value := range paths {
		rel, err := filepath.Rel(filepath.Dir(savePath), value)
		utils.HandleError(err)
		fileName := filepath.Base(value)
		file += " " + fileName[0:len(fileName)-len(filepath.Ext(fileName))] + " in '" + rel + "'"
		if key == len(paths)-1 {
			file += ";\n\n\n"
		} else {
			file += ",\n"
		}
	}
	file += "end."

	err := ioutil.WriteFile(savePath, []byte(file), os.ModePerm)
	utils.HandleError(err)
	return err == nil
}

func generateDproj(dpkName string) {
	var baseName = filepath.Base(dpkName)
	var name = baseName[0 : len(baseName)-len(filepath.Ext(baseName))]

	var file string
	file += "<Project xmlns=\"http://schemas.microsoft.com/developer/msbuild/2003\">\n"
	file += "    <PropertyGroup>\n"
	file += "        <ProjectGuid>{820BE041-A04E-4B7D-B976-50A82AEF3AB9}</ProjectGuid>\n"
	file += "        <MainSource>" + baseName + "</MainSource>\n"
	file += "        <Base>True</Base>\n"
	file += "        <Config Condition=\"'$(Config)'==''\">Debug</Config>\n"
	file += "        <TargetedPlatforms>129</TargetedPlatforms>\n"
	file += "        <AppType>Package</AppType>\n"
	file += "        <FrameworkType>None</FrameworkType>\n"
	file += "        <ProjectVersion>18.7</ProjectVersion>\n"
	file += "        <Platform Condition=\"'$(Platform)'==''\">Win32</Platform>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Config)'=='Base' or '$(Base)'!=''\">\n"
	file += "        <Base>true</Base>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"('$(Platform)'=='Android' and '$(Base)'=='true') or '$(Base_Android)'!=''\">\n"
	file += "        <Base_Android>true</Base_Android>\n"
	file += "        <CfgParent>Base</CfgParent>\n"
	file += "        <Base>true</Base>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"('$(Platform)'=='Win32' and '$(Base)'=='true') or '$(Base_Win32)'!=''\">\n"
	file += "        <Base_Win32>true</Base_Win32>\n"
	file += "        <CfgParent>Base</CfgParent>\n"
	file += "        <Base>true</Base>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Config)'=='Release' or '$(Cfg_1)'!=''\">\n"
	file += "        <Cfg_1>true</Cfg_1>\n"
	file += "        <CfgParent>Base</CfgParent>\n"
	file += "        <Base>true</Base>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Config)'=='Debug' or '$(Cfg_2)'!=''\">\n"
	file += "        <Cfg_2>true</Cfg_2>\n"
	file += "        <CfgParent>Base</CfgParent>\n"
	file += "        <Base>true</Base>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Base)'!=''\">\n"
	file += "        <DCC_E>false</DCC_E>\n"
	file += "        <DCC_F>false</DCC_F>\n"
	file += "        <DCC_K>false</DCC_K>\n"
	file += "        <DCC_N>false</DCC_N>\n"
	file += "        <DCC_S>false</DCC_S>\n"
	file += "        <DCC_ImageBase>00400000</DCC_ImageBase>\n"
	file += "        <GenDll>true</GenDll>\n"
	file += "        <GenPackage>true</GenPackage>\n"
	file += "        <SanitizedProjectName>" + name + "</SanitizedProjectName>\n"
	file += "        <VerInfo_Locale>1046</VerInfo_Locale>\n"
	file += "        <VerInfo_Keys>CompanyName=;FileDescription=;FileVersion=1.0.0.0;InternalName=;LegalCopyright=;LegalTrademarks=;OriginalFilename=;ProductName=;ProductVersion=1.0.0.0;Comments=;CFBundleName=</VerInfo_Keys>\n"
	file += "        <DCC_Namespace>System;Xml;Data;Datasnap;Web;Soap;$(DCC_Namespace)</DCC_Namespace>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Base_Android)'!=''\">\n"
	file += "        <VerInfo_Keys>package=com.embarcadero.$(MSBuildProjectName);label=$(MSBuildProjectName);versionCode=1;versionName=1.0.0;persistent=False;restoreAnyVersion=False;installLocation=auto;largeHeap=False;theme=TitleBar;hardwareAccelerated=true;apiKey=</VerInfo_Keys>\n"
	file += "        <BT_BuildType>Debug</BT_BuildType>\n"
	file += "        <EnabledSysJars>android-support-v4.dex.jar;cloud-messaging.dex.jar;com-google-android-gms.play-services-ads-base.17.2.0.dex.jar;com-google-android-gms.play-services-ads-identifier.16.0.0.dex.jar;com-google-android-gms.play-services-ads-lite.17.2.0.dex.jar;com-google-android-gms.play-services-ads.17.2.0.dex.jar;com-google-android-gms.play-services-analytics-impl.16.0.8.dex.jar;com-google-android-gms.play-services-analytics.16.0.8.dex.jar;com-google-android-gms.play-services-base.16.0.1.dex.jar;com-google-android-gms.play-services-basement.16.2.0.dex.jar;com-google-android-gms.play-services-gass.17.2.0.dex.jar;com-google-android-gms.play-services-identity.16.0.0.dex.jar;com-google-android-gms.play-services-maps.16.1.0.dex.jar;com-google-android-gms.play-services-measurement-base.16.4.0.dex.jar;com-google-android-gms.play-services-measurement-sdk-api.16.4.0.dex.jar;com-google-android-gms.play-services-stats.16.0.1.dex.jar;com-google-android-gms.play-services-tagmanager-v4-impl.16.0.8.dex.jar;com-google-android-gms.play-services-tasks.16.0.1.dex.jar;com-google-android-gms.play-services-wallet.16.0.1.dex.jar;com-google-firebase.firebase-analytics.16.4.0.dex.jar;com-google-firebase.firebase-common.16.1.0.dex.jar;com-google-firebase.firebase-iid-interop.16.0.1.dex.jar;com-google-firebase.firebase-iid.17.1.1.dex.jar;com-google-firebase.firebase-measurement-connector.17.0.1.dex.jar;com-google-firebase.firebase-messaging.17.5.0.dex.jar;fmx.dex.jar;google-play-billing.dex.jar;google-play-licensing.dex.jar</EnabledSysJars>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Base_Win32)'!=''\">\n"
	file += "        <DCC_Namespace>Winapi;System.Win;Data.Win;Datasnap.Win;Web.Win;Soap.Win;Xml.Win;Bde;$(DCC_Namespace)</DCC_Namespace>\n"
	file += "        <BT_BuildType>Debug</BT_BuildType>\n"
	file += "        <VerInfo_IncludeVerInfo>true</VerInfo_IncludeVerInfo>\n"
	file += "        <VerInfo_Keys>CompanyName=;FileDescription=$(MSBuildProjectName);FileVersion=1.0.0.0;InternalName=;LegalCopyright=;LegalTrademarks=;OriginalFilename=;ProductName=$(MSBuildProjectName);ProductVersion=1.0.0.0;Comments=;ProgramID=com.embarcadero.$(MSBuildProjectName)</VerInfo_Keys>\n"
	file += "        <VerInfo_Locale>1033</VerInfo_Locale>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Cfg_1)'!=''\">\n"
	file += "        <DCC_Define>RELEASE;$(DCC_Define)</DCC_Define>\n"
	file += "        <DCC_DebugInformation>0</DCC_DebugInformation>\n"
	file += "        <DCC_LocalDebugSymbols>false</DCC_LocalDebugSymbols>\n"
	file += "        <DCC_SymbolReferenceInfo>0</DCC_SymbolReferenceInfo>\n"
	file += "    </PropertyGroup>\n"
	file += "    <PropertyGroup Condition=\"'$(Cfg_2)'!=''\">\n"
	file += "        <DCC_Define>DEBUG;$(DCC_Define)</DCC_Define>\n"
	file += "        <DCC_Optimize>false</DCC_Optimize>\n"
	file += "        <DCC_GenerateStackFrames>true</DCC_GenerateStackFrames>\n"
	file += "    </PropertyGroup>\n"
	file += "    <ItemGroup>\n"
	file += "        <DelphiCompile Include=\"$(MainSource)\">\n"
	file += "            <MainSource>MainSource</MainSource>\n"
	file += "        </DelphiCompile>\n"
	file += "        <BuildConfiguration Include=\"Debug\">\n"
	file += "            <Key>Cfg_2</Key>\n"
	file += "            <CfgParent>Base</CfgParent>\n"
	file += "        </BuildConfiguration>\n"
	file += "        <BuildConfiguration Include=\"Base\">\n"
	file += "            <Key>Base</Key>\n"
	file += "        </BuildConfiguration>\n"
	file += "        <BuildConfiguration Include=\"Release\">\n"
	file += "            <Key>Cfg_1</Key>\n"
	file += "            <CfgParent>Base</CfgParent>\n"
	file += "        </BuildConfiguration>\n"
	file += "    </ItemGroup>\n"
	file += "    <ProjectExtensions>\n"
	file += "        <Borland.Personality>Delphi.Personality.12</Borland.Personality>\n"
	file += "        <Borland.ProjectType>Package</Borland.ProjectType>\n"
	file += "        <BorlandProject>\n"
	file += "            <Delphi.Personality>\n"
	file += "                <Source>\n"
	file += "                    <Source Name=\"MainSource\">" + baseName + "</Source>\n"
	file += "                </Source>\n"
	file += "            </Delphi.Personality>\n"
	file += "            <Platforms>\n"
	file += "                <Platform value=\"Android\">False</Platform>\n"
	file += "                <Platform value=\"iOSDevice32\">False</Platform>\n"
	file += "                <Platform value=\"iOSSimulator\">False</Platform>\n"
	file += "                <Platform value=\"Linux64\">True</Platform>\n"
	file += "                <Platform value=\"OSX32\">False</Platform>\n"
	file += "                <Platform value=\"Win32\">True</Platform>\n"
	file += "                <Platform value=\"Win64\">False</Platform>\n"
	file += "            </Platforms>\n"
	file += "        </BorlandProject>\n"
	file += "        <ProjectFileVersion>12</ProjectFileVersion>\n"
	file += "    </ProjectExtensions>\n"
	file += "    <Import Project=\"$(BDS)\\Bin\\CodeGear.Delphi.Targets\" Condition=\"Exists('$(BDS)\\Bin\\CodeGear.Delphi.Targets')\"/>\n"
	file += "    <Import Project=\"$(APPDATA)\\Embarcadero\\$(BDSAPPDATABASEDIR)\\$(PRODUCTVERSION)\\UserTools.proj\" Condition=\"Exists('$(APPDATA)\\Embarcadero" +
		"\\$(BDSAPPDATABASEDIR)\\$(PRODUCTVERSION)\\UserTools.proj')\"/>\n"
	file += "</Project>\n"
	file += "\n"

	err := ioutil.WriteFile(filepath.Join(filepath.Dir(dpkName), name+consts.FileExtensionDproj), []byte(file), os.ModePerm)
	utils.HandleError(err)

}
