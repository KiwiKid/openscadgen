include <BOSL2/std.scad>;

$fn = 100;

standardHexSize = 6.35/2;

holderHexSize = standardHexSize;

holderHexLength = 30;
holderHexLength2 = 50;

clipHexDownAdjust=  0.5;
clipHexOutAdjust = -2;
housingHolderUp = 2;

holderHexHoleSize = standardHexSize+0.1;
holderHexHoleHeight = 20;

holeCount = 7;

holeSeperator = holderHexHoleSize/2+2.5;
holeEndOffset =5.1;

holeUp = 0 + housingHolderUp;
holeOutFromClip = holderHexSize*1.3+0.5;
holeMove = [holderHexLength2/2-holeEndOffset, -holeOutFromClip, holeUp];

holeSpin = -2;
holeLean = 14;
holeRotate = [-holeLean,30,0];


holeHelperMove = [-2,-2, -3];

globalBounds = [100,100,standardHexSize*2];

module multi_hex_housing_part(hexLength=10, hexRadius=standardHexSize){
    yrot(90)
    linear_extrude(height = hexLength)
    hexagon(r=hexRadius, anchor=CENTER);
    }
    
    
 connectorSeperator1 = 1.3;
  connectorSeperator2 = 0.3;
  connectorUp = 3;

module multi_hex_housing(){
    //left(holderHexLength/2
             up(connectorUp)

    
    //xcyl(r=standardHexSize*2, h= holderHexLength*0.5, anchor=CENTER);
    hull(){
    
    
    /* right((holderHexLength2-holderHexLength)/2){
     /* back(holderHexSize*connectorSeperator1)
        multi_hex_housing_part(hexLength= holderHexLength);
        

        

        
      //  fwd(holderHexSize*connectorSeperator2)
       // right(10)
      ///  back(3)
      back(0)
        right(10)
        xrot(-30)
            multi_hex_housing_part(hexLength= holderHexLength-20, hexRadius=5);
        
        }*/
        right(holderHexLength/2)
        right((holderHexLength2-holderHexLength)/2){
        cuboid([15,10,2]);
        }
         up(housingHolderUp)
        fwd(holderHexSize*connectorSeperator2)
       multi_hex_housing_part(hexLength= holderHexLength2);
            
        }
        
       
        xrot(30)
        down(clipHexDownAdjust)
        fwd(clipHexOutAdjust)
              right((holderHexLength2-holderHexLength)/2)
    back(holderHexSize*connectorSeperator1)
        multi_hex_housing_part(hexLength= holderHexLength);
        

    
}
module multi_hex_holes(mode="isHoles", holderHexRadius = holderHexHoleSize){

    xcopies(holderHexHoleSize+holeSeperator,holeCount)
    move(holeMove)
    rotate(holeRotate){
    linear_extrude(height = holderHexHoleHeight)
    hexagon(r=holderHexRadius, anchor=CENTER);
       
       if(mode == "isHoles"){

       move(holeHelperMove)
       cyl(r=holderHexSize*.7, h=10);
       }
    }
}


module multi_hex_drill_attachment(){
intersection(){
    difference(){
    union(){
    
        multi_hex_housing();
    /*    down(11)
        back(1)
        right(0.7)
        left_half()
        multi_hex_holes(mode="isNotHoles", holderHexRadius = holderHexHoleSize+1);*/
        }
        down(7)
        multi_hex_holes(mode="isHoles");
    }
    #cuboid(globalBounds, anchor=BOT);
    }
}
multi_hex_drill_attachment();
