

	include <BOSL2/std.scad>;
	include <BOSL2/joiners.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";


    isSizer = "true";
    sizerCount = 20;
    
    onlySection = "all";
    
    dovetailHeight = 25;
    
    dovetailWidth = 25;
    
    dovetailSlide = 3;
    
    
    sizerWidth = 35;
    sizerLength = 120;
    sizerHeight = 3;
    
    sizerOffset = 5;
    
    
    holderSize = [40,400,10];
    
        oilWidth=10;

    oilHeight=sizerLength*2-40;
    
    oilSize = [oilWidth,oilHeight,2];
    
    oilPickUp = [5,10,5];
    
    sliderOffset = 28;
    
    module dovetailSection(section="all", dovetailHeight=dovetailHeight, dovetailWidth=dovetailWidth, dovetailSlide=dovetailSlide){
    diff("remove")
       down(sizerHeight/2-1)
       rotate([0,0,90])
       fwd(10)
      // up(10)
      right(4)
        cuboid([sizerWidth,sizerLength,sizerHeight], anchor=CENTER, rounding=1, edges="Z"){
            if(section == "all" || section == "male"){
            attach(BACK) 
            dovetail("male", slide=dovetailSlide, width=dovetailWidth, height=dovetailHeight, radius=0.5, round=true);
            }
             if(section == "all" || section == "female"){
            tag("remove")
            attach(FRONT) 
            dovetail("female", slide=dovetailSlide, width=dovetailWidth, height=dovetailHeight, radius=0.5, round=true);
            }
            }
            
            }

    module sizerSection(){
    
        difference(){

        
   /*    
            
            
            
        }*/
        
        fwd(14)
        move([12,12,1])
        rotate([0,0,180])
        text3d(str(sizerCount-$idx), h=1, size=6);
        }
        
        
     difference(){   
           fwd(5)
           left(1)
          cuboid([5,1,2]);
        }
        }
    
    
    
    titleLocation = [11.5,-sizerLength+13,1];
    
    coverSize = [13.5,26.7,10];
    coverLocation = [0,0,-1.5];
    
    
        diceLocation = [-10,-sizerLength+28,7.20];
    diceSize = 17.5;


	module oil_counter(){
    
        difference(){
        
        union(){
        //cuboid(holderSize);
        
      //  up(4)
        // ycopies(sizerLength, n=2){
        //     if(onlySection == str(2-$idx) || onlySection == "all"){

        if(onlySection == "1" || onlySection == "all"){
            back(sizerLength/2-sizerOffset)
            rotate([0,0,90])
            dovetailSection(section="male");
          }
           if(onlySection == "2" || onlySection == "all"){
          
           fwd(sizerLength/2+sizerOffset)
              rotate([0,0,90])
            dovetailSection(section="female");
            }
            
            
        }
        
    
        up(1)
        left(10)
        back(sliderOffset)
        fwd(sizerOffset)
        cuboid(oilSize, rounding=-2, edges=TOP);
        
        
            
        move(diceLocation)
        cuboid(diceSize, rounding=3, edges=BOTTOM);
        
        move(diceLocation+coverLocation)
        cuboid(coverSize, edges=BOTTOM);
        
        
		up(0.1)
        back(sliderOffset)
        ycopies(10, n=sizerCount){
        
            sizerSection();
            
            
            }
            
        move(titleLocation)
        rotate([0,0,180])
        text3d("Propane", h=1, size=6);
        
        }
	}


    sliced(renderType=renderType) {
        oil_counter();
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

