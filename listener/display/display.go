package display




/*
#include <windows.h>
#include <stdio.h>

void
SetResolution(char* display_name,
			  int Width,
			  int Height) {
	HRESULT result = 1;
	int deviceIndex = 0;
	do
	{
		DISPLAY_DEVICE dpd = {0};
		PDISPLAY_DEVICE displayDevice = &dpd;
		displayDevice->cb = sizeof(DISPLAY_DEVICE);

		result = EnumDisplayDevices(NULL, 
			deviceIndex++, displayDevice, 0);
		if (displayDevice->StateFlags & DISPLAY_DEVICE_ACTIVE)
		{
			PDISPLAY_DEVICE monitor = (PDISPLAY_DEVICE)malloc(sizeof(DISPLAY_DEVICE));
			monitor->cb = sizeof(DISPLAY_DEVICE);

            if(strcmp(displayDevice->DeviceName,display_name))
                continue;

			EnumDisplayDevices(displayDevice->DeviceName, 
				0, monitor, 0);
			
			PDEVMODE dm = (PDEVMODE)malloc(sizeof(DEVMODE));
			if ( EnumDisplaySettings(displayDevice->DeviceName, ENUM_CURRENT_SETTINGS, dm) ) {
                dm->dmPelsWidth  = Width;
                dm->dmPelsHeight = Height;
                printf("changing resolution of display %ls to %dx%d \n",displayDevice->DeviceString,Width,Height);
				ChangeDisplaySettingsEx(displayDevice->DeviceName, dm,  \
                                     NULL, (CDS_GLOBAL | CDS_UPDATEREGISTRY | CDS_RESET), NULL);
			}
		}
	} while (result);
}



typedef struct {
	char display_names[100][100];
}Displays;

Displays*
GetDisplay() {
	static Displays displays = {0};
	memset(&displays,0,sizeof(Displays));

	HRESULT result = 1;
	int deviceIndex = 0;
	int resultIndex = 0;
	do
	{
		DISPLAY_DEVICE dpd = {0};
		PDISPLAY_DEVICE displayDevice = &dpd;
		displayDevice->cb = sizeof(DISPLAY_DEVICE);

		result = EnumDisplayDevices(NULL, 
			deviceIndex, displayDevice, 0);
		if (displayDevice->StateFlags & DISPLAY_DEVICE_ACTIVE)
		{
			PDISPLAY_DEVICE monitor = (PDISPLAY_DEVICE)malloc(sizeof(DISPLAY_DEVICE));
			monitor->cb = sizeof(DISPLAY_DEVICE);

			EnumDisplayDevices(displayDevice->DeviceName, 
				0, monitor, 0);
			
			memcpy(displays.display_names[resultIndex],displayDevice->DeviceName,strlen(displayDevice->DeviceName));
			resultIndex++;
		}
		deviceIndex++;
	} while (result);

    return &displays;
}


*/
import "C"


var (
	current_display = ""
)

func SetResolution( DisplayName string,
					width int,
					height int,
					) {
	C.SetResolution(C.CString(DisplayName),C.int(width),C.int(height))
}

func GetDisplays() []string{
	result := []string{}
	displays := C.GetDisplay()
	for _,n := range displays.display_names {
		result = append(result, C.GoString(&n[0]))
	}
	return result
}